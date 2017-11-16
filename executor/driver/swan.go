package driver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	uuid "github.com/satori/go.uuid"

	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/Dataman-Cloud/swan/utils/recordio"
)

// SwanExecDriver is an implement of Driver interface
type SwanDriver struct {
	sync.RWMutex
	cfg    *DriverConfig
	status *mesosproto.Status // DRIVER_NOT_STARTED, DRIVER_RUNNING, DRIVER_ABORTED, DRIVER_STOPPED
	client *http.Client

	stopCh chan struct{}

	// event handlers
	handlers map[mesosproto.ExecEvent_Type][]EventHandler
}

func NewSwanDriver() (*SwanDriver, error) {
	log.Println("initializing swan mesos executor driver ...")

	driver := new(SwanDriver)

	// obtian configs from environments
	cfg, err := NewDriverConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("NewDriverConfigFromEnv() error: %v", err)
	}
	driver.cfg = cfg

	log.Println("obtained mesos executor configs:", cfg)

	// init http client
	driver.client = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
		},
	}

	// init handlers map
	driver.handlers = make(map[mesosproto.ExecEvent_Type][]EventHandler)

	// init stop channel
	driver.stopCh = make(chan struct{})

	return driver, nil
}

func (d *SwanDriver) RegisterEventHandlers(evType mesosproto.ExecEvent_Type, handleFuncs []EventHandler) {
	d.Lock()
	defer d.Unlock()
	d.handlers[evType] = handleFuncs
}

func (d *SwanDriver) Start() error {
	log.Println("swan executor driver starting")
	err := d.subscribe()
	if err != nil {
		return err
	}
	log.Println("swan executor driver started")

	<-d.stopCh

	log.Println("swan executor driver stopped")
	return nil
}

func (d *SwanDriver) Stop() {
	log.Println("swan executor driver stopping")
	close(d.stopCh)
}

func (d *SwanDriver) labels() *mesosproto.Labels {
	return &mesosproto.Labels{
		Labels: []*mesosproto.Label{
			&mesosproto.Label{
				Key:   proto.String("MESOS_EXECUTOR_DRIVER_NAME"),
				Value: proto.String("swan"),
			},
		},
	}
}

func (d *SwanDriver) SendStatusUpdate(taskId *mesosproto.TaskID, state *mesosproto.TaskState, message *string) error {
	var (
		uuid     = uuid.NewV4().Bytes()
		uuidText = base64.StdEncoding.EncodeToString(uuid)
	)

	err := d.sendUpdate(&mesosproto.TaskStatus{
		TaskId:     taskId,
		State:      state,
		Message:    message,
		Source:     mesosproto.TaskStatus_SOURCE_EXECUTOR.Enum(),
		AgentId:    &mesosproto.AgentID{Value: proto.String(d.cfg.slaveID)},
		ExecutorId: d.cfg.executorID,
		Timestamp:  proto.Float64(float64(time.Now().Unix())),
		Uuid:       []byte(uuid),
		Labels:     d.labels(),
	})

	if err != nil {
		log.Errorf("SendStatusUpdate() taskid=%s, state=%s error: %v", taskId.GetValue(), state.String(), err)
	} else {
		log.Printf("SendStatusUpdate() taskid=%s, state=%s, uuid=%s succeed", taskId.GetValue(), state.String(), uuidText)
	}

	return err
}

func (d *SwanDriver) SendFrameworkMessage(msg string) error {
	call := &mesosproto.ExecCall{
		Type:        mesosproto.ExecCall_MESSAGE.Enum(),
		FrameworkId: d.cfg.frameworkID,
		ExecutorId:  d.cfg.executorID,
		Message: &mesosproto.ExecCall_Message{
			Data: []byte(msg),
		},
	}

	resp, err := d.SendCall(call, http.StatusAccepted)
	if err != nil {
		log.Errorf("SendFrameworkMessage() [%s] error: %v", msg, err)
		return err
	}

	log.Printf("SendFrameworkMessage() [%s] succeed", msg)
	resp.Body.Close()
	return nil
}

func (d *SwanDriver) EndPoint() string {
	return "http://" + d.cfg.slaveEndPoint + MesosExecutorApiEndPoint
}

// Utils
//
//
func (d *SwanDriver) subscribe() error {
	call := &mesosproto.ExecCall{
		Type:        mesosproto.ExecCall_SUBSCRIBE.Enum(),
		FrameworkId: d.cfg.frameworkID,
		ExecutorId:  d.cfg.executorID,
		Subscribe:   new(mesosproto.ExecCall_Subscribe),
	}

	resp, err := d.SendCall(call, http.StatusOK)
	if err != nil {
		return err
	}

	go d.handleEvents(resp.Body)
	return nil
}

func (d *SwanDriver) sendUpdate(status *mesosproto.TaskStatus) error {
	call := &mesosproto.ExecCall{
		Type:        mesosproto.ExecCall_UPDATE.Enum(),
		FrameworkId: d.cfg.frameworkID,
		ExecutorId:  d.cfg.executorID,
		Update: &mesosproto.ExecCall_Update{
			Status: status,
		},
	}

	resp, err := d.SendCall(call, http.StatusAccepted)
	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}

// TODO Handle Disconnections and Retry Logic and  re-Subscribe Unacknowledged Tasks & Updates
func (d *SwanDriver) handleEvents(stream io.ReadCloser) {

	var (
		ev  *mesosproto.ExecEvent
		err error
	)

	dec := json.NewDecoder(recordio.NewReader(stream))

	// decode & handle each event by order
	for {
		if err = dec.Decode(&ev); err != nil {
			log.Errorf("mesos events subscriber decode events error: %v", err)
			stream.Close()
			return
		}

		if err := d.handleEvent(ev); err != nil {
			log.Errorf("executor handle event [%s] error [%v]", ev.GetType(), err)
		}
	}
}

func (d *SwanDriver) handleEvent(ev *mesosproto.ExecEvent) error {
	var (
		typ           = ev.GetType()
		evHandleFuncs = d.handlers[typ]
	)

	if len(evHandleFuncs) == 0 {
		return fmt.Errorf("without any proper event handler for mesos event: %v", typ)
	}

	for _, evHandleFunc := range evHandleFuncs {
		err := evHandleFunc(d, ev)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *SwanDriver) ackUpdateEvent(status *mesosproto.TaskStatus) error {
	return errors.New("not implemented yet")
}

// SendCall send mesos request against the mesos master's scheduler api endpoint.
// note it's the caller's responsibility to deal with the SendCall() error
func (d *SwanDriver) SendCall(call *mesosproto.ExecCall, expectCode int) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, fmt.Errorf("sendCall().Marshal() error %v", err)
	}

	resp, err := d.send(payload)
	if err != nil {
		return nil, fmt.Errorf("sendCall().send() error %v", err)
	}

	if code := resp.StatusCode; code != expectCode {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("sendCall() with unexpected response [%d] - [%s]", code, string(bs))
	}

	return resp, nil
}

func (d *SwanDriver) send(payload []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("POST", d.EndPoint(), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := d.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request: %s", err)
	}

	return httpResp, nil
}
