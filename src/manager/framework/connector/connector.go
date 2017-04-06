package connector

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/utils"

	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
	"github.com/golang/protobuf/proto"
	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

var SPECIAL_CHARACTER = regexp.MustCompile("([\\-\\.\\$\\*\\+\\?\\{\\}\\(\\)\\[\\]\\|]+)")

var instance *Connector
var once sync.Once

type Connector struct {
	ClusterID             string
	MesosLeader           string
	MesosLeaderHttpClient *HttpClient

	SendChan    chan *sched.Call
	ReceiveChan chan *event.MesosEvent

	mesosFailureChan chan error

	FrameworkInfo *mesos.FrameworkInfo

	StreamCtx       context.Context
	StreamCancelFun context.CancelFunc
}

func NewConnector() *Connector {
	return Instance() // call initialize method
}

func Instance() *Connector {
	once.Do(
		func() {
			hostname, _ := os.Hostname()
			info := &mesos.FrameworkInfo{
				User:      proto.String(config.SchedulerConfig.MesosFrameworkUser),
				Name:      proto.String("swan"),
				Principal: proto.String("swan"),

				FailoverTimeout: proto.Float64(60 * 60 * 3),
				Checkpoint:      proto.Bool(false),
				Hostname:        proto.String(hostname),
				Capabilities: []*mesos.FrameworkInfo_Capability{
					&mesos.FrameworkInfo_Capability{Type: mesos.FrameworkInfo_Capability_PARTITION_AWARE.Enum()},
					&mesos.FrameworkInfo_Capability{Type: mesos.FrameworkInfo_Capability_TASK_KILLING_STATE.Enum()},
				},
			}

			instance = &Connector{
				ReceiveChan:   make(chan *event.MesosEvent, 1024), // make this unbound in future
				SendChan:      make(chan *sched.Call, 1024),
				FrameworkInfo: info,
			}
		})

	return instance
}

func (s *Connector) Subscribe(ctx context.Context) {
	logrus.Infof("subscribe to mesos leader: %s", s.MesosLeader)

	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: s.FrameworkInfo,
		},
	}

	if s.FrameworkInfo.Id != nil {
		call.FrameworkId = &mesos.FrameworkID{
			Value: proto.String(s.FrameworkInfo.Id.GetValue()),
		}
	}

	resp, err := s.Send(call)
	if err != nil {
		logrus.Errorf("send subscribe call got err: %d", err)
		s.mesosFailureChan <- utils.NewError(utils.SeverityLow, err)

		logrus.Error("exiting Subscribe")
		return
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("subscribe got http response status code: %d", resp.StatusCode)
		s.mesosFailureChan <- utils.NewError(utils.SeverityLow,
			errors.New(fmt.Sprintf("subscribe with unexpected response status: %d", resp.StatusCode)))

		logrus.Error("exiting Subscribe")
		return
	}

	s.handleEvents(ctx, resp)
}

func (s *Connector) handleEvents(ctx context.Context, resp *http.Response) {
	defer func() {
		resp.Body.Close()
	}()

	r := NewReader(resp.Body)
	dec := json.NewDecoder(r)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("goroutine handleEvents cancelled %s", ctx.Err())
			return
		default:
			event := new(sched.Event)
			if err := dec.Decode(event); err != nil {
				logrus.Errorf("handleEvents goroutine decode response got err: %s", err)
				s.mesosFailureChan <- utils.NewError(utils.SeverityLow, err)

				logrus.Error("goroutine handleEvents exiting")
				return
			}

			switch event.GetType() {
			case sched.Event_SUBSCRIBED:
				logrus.Infof("subscribed successful with ID %s", event.GetSubscribed().FrameworkId.GetValue())
				s.addEvent(sched.Event_SUBSCRIBED, event)
			case sched.Event_OFFERS:
				s.addEvent(sched.Event_OFFERS, event)
			case sched.Event_RESCIND:
				s.addEvent(sched.Event_RESCIND, event)
			case sched.Event_UPDATE:
				s.addEvent(sched.Event_UPDATE, event)
			case sched.Event_MESSAGE:
				s.addEvent(sched.Event_MESSAGE, event)
			case sched.Event_FAILURE:
				s.addEvent(sched.Event_FAILURE, event)
			case sched.Event_ERROR:
				s.addEvent(sched.Event_ERROR, event)
			case sched.Event_HEARTBEAT:
				s.addEvent(sched.Event_HEARTBEAT, event)
			}
		}
	}
}

func (s *Connector) SetFrameworkInfoId(id string) {
	s.FrameworkInfo.Id = &mesos.FrameworkID{Value: proto.String(id)}
}

func getMastersFromZK(zkPath string) ([]string, error) {
	masterInfo := new(mesos.MasterInfo)

	if !strings.HasPrefix(zkPath, "zk://") {
		zkPath = fmt.Sprintf("zk://%s", zkPath)
	}
	url, err := url.Parse(zkPath)
	if err != nil {
		return nil, err
	}

	conn, _, err := zk.Connect(strings.Split(url.Host, ","), 5*time.Second)
	defer conn.Close()
	if err != nil {
		return nil, err
	}

	children, _, err := conn.Children(url.Path)
	if err != nil {
		return nil, err
	}

	masters := make([]string, 0)
	for _, node := range children {
		if strings.HasPrefix(node, "json.info") {
			data, _, _ := conn.Get(url.Path + "/" + node)
			err := json.Unmarshal(data, masterInfo)
			if err != nil {
				return nil, err
			}
			masters = append(masters, fmt.Sprintf("%s:%d", *masterInfo.GetAddress().Ip, *masterInfo.GetAddress().Port))
		}
	}

	return masters, nil
}

func stateFromMasters(masters []string) (*megos.State, error) {
	masterUrls := make([]*url.URL, 0)
	for _, master := range masters {
		masterUrl, _ := url.Parse(fmt.Sprintf("http://%s", master))
		masterUrls = append(masterUrls, masterUrl)
	}

	mesos := megos.NewClient(masterUrls, nil)
	return mesos.GetStateFromCluster()
}

func (s *Connector) Send(call *sched.Call) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}
	return s.MesosLeaderHttpClient.Send(payload)
}

func (s *Connector) addEvent(eventType sched.Event_Type, e *sched.Event) {
	s.ReceiveChan <- &event.MesosEvent{EventType: eventType, Event: e}
}

func (s *Connector) Reregister() error {
	logrus.Infof("register to mesos now")

	if s.StreamCancelFun != nil {
		s.StreamCancelFun()
	}

	err := s.leaderDetect()
	if err != nil { // if leader detect encounter any error
		logrus.Errorf("exiting reregister due to err: %s", err)
		return err
	}

	s.StreamCtx, s.StreamCancelFun = context.WithCancel(context.Background())
	go s.Subscribe(s.StreamCtx)
	return nil
}

func (s *Connector) Start(ctx context.Context, errorChan chan error) {
	s.mesosFailureChan = errorChan
	err := s.leaderDetect()
	if err != nil {
		logrus.Errorf("start mesos connector got error: %s", err)
		s.mesosFailureChan <- utils.NewError(utils.SeverityHigh, err) // set SeverityHigh when first start
		return
	}

	s.StreamCtx, s.StreamCancelFun = context.WithCancel(context.Background())
	go s.Subscribe(s.StreamCtx)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("connector got done signal: %s", ctx.Err())
			s.StreamCancelFun() // stop stream goroutine
			return
		case call := <-s.SendChan:
			//logrus.WithFields(logrus.Fields{"sending-call": sched.Call_Type_name[int32(*call.Type)]}).Debugf("%+v", call)
			resp, err := s.Send(call)
			if err != nil {
				logrus.Errorf("send call to master got err: %s", err)
				s.mesosFailureChan <- utils.NewError(utils.SeverityLow, err)
			}
			if resp != nil && resp.StatusCode != 202 {
				logrus.Errorf("send call to master response not valid: %d", resp.StatusCode)
				s.mesosFailureChan <- utils.NewError(utils.SeverityLow, errors.New("sending call response status code not 202"))
			}
		}
	}
}

func (s *Connector) leaderDetect() error {
	masters, err := getMastersFromZK(config.SchedulerConfig.ZkPath)
	if err != nil {
		return err
	}

	state, err := stateFromMasters(masters)
	if err != nil {
		return err
	}

	s.MesosLeaderHttpClient = NewHTTPClient(state.Leader, "/api/v1/scheduler")
	s.MesosLeader = state.Leader

	if len(strings.TrimSpace(state.Cluster)) == 0 {
		s.ClusterID = "cluster"
	} else {
		s.ClusterID = state.Cluster
	}

	if SPECIAL_CHARACTER.MatchString(s.ClusterID) {
		logrus.Warnf(`Swan do not work with mesos cluster name(%s) with special characters "-.$*+?{}()[]|".`, s.ClusterID)
		s.ClusterID = SPECIAL_CHARACTER.ReplaceAllString(s.ClusterID, "")
	}

	return nil
}
