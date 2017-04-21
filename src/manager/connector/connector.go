package connector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/event"
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

// Connector create a persistent connection against mesos master to subscribe mesos event.
// caller could use:
// MesosEvent() to subscribe mesos events
// ErrEvent()   to subscribe connnector's failures
// SendCall()   to emit request against mesos master
type Connector struct {
	MesosZkPath           *url.URL
	ClusterID             string
	MesosLeader           string
	MesosLeaderHttpClient *HttpClient

	EventChan chan *event.MesosEvent
	ErrorChan chan error

	FrameworkInfo *mesos.FrameworkInfo

	StreamCtx       context.Context
	StreamCancelFun context.CancelFunc
}

func Instance() *Connector {
	return instance
}

func Init(user string, mesosZkPath *url.URL) {
	once.Do(
		func() {
			hostname, _ := os.Hostname()
			info := &mesos.FrameworkInfo{
				User:      proto.String(user),
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
				MesosZkPath:   mesosZkPath,
				EventChan:     make(chan *event.MesosEvent, 1024),
				ErrorChan:     make(chan error, 1024),
				FrameworkInfo: info,
			}
		})
}

func (s *Connector) subscribe(ctx context.Context) {
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

	resp, err := s.send(call)
	if err != nil {
		logrus.Errorf("send subscribe call got err: %v, abort", err)
		s.emitError(utils.SeverityLow, err)
		return
	}

	if code := resp.StatusCode; code != http.StatusOK {
		logrus.Errorf("subscribe expect 200, got %d, abort", code)
		s.emitError(utils.SeverityLow, fmt.Sprintf("subscribe with unexpected response status: %d", code))
		return
	}

	s.handleEvents(ctx, resp)
}

func (s *Connector) handleEvents(ctx context.Context, resp *http.Response) {
	defer resp.Body.Close()

	r := NewReader(resp.Body)
	dec := json.NewDecoder(r)

	for {
		select {

		case <-ctx.Done():
			logrus.Infof("goroutine handleEvents cancelled %v", ctx.Err())
			return

		default:
			event := new(sched.Event)
			if err := dec.Decode(event); err != nil {
				logrus.Errorf("handleEvents goroutine decode response got err: %v, abort", err)
				s.emitError(utils.SeverityLow, err)
				return
			}

			s.emitEvent(event.GetType(), event)
		}
	}
}

func (s *Connector) SetFrameworkInfoId(id string) {
	s.FrameworkInfo.Id = &mesos.FrameworkID{Value: proto.String(id)}
}

func getMastersFromZK(zkPath *url.URL) ([]string, error) {
	masterInfo := new(mesos.MasterInfo)
	conn, _, err := zk.Connect(strings.Split(zkPath.Host, ","), 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	children, _, err := conn.Children(zkPath.Path)
	if err != nil {
		return nil, err
	}

	masters := make([]string, 0)
	for _, node := range children {
		if strings.HasPrefix(node, "json.info") {
			data, _, err := conn.Get(zkPath.Path + "/" + node)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(data, masterInfo); err != nil {
				return nil, err
			}
			masters = append(masters, fmt.Sprintf("%s:%d",
				*masterInfo.GetAddress().Ip, *masterInfo.GetAddress().Port))
		}
	}

	return masters, nil
}

func stateFromMasters(masters []string) (*megos.State, error) {
	masterUrls := make([]*url.URL, 0, len(masters))
	for _, master := range masters {
		masterUrl, _ := url.Parse(fmt.Sprintf("http://%s", master))
		masterUrls = append(masterUrls, masterUrl)
	}

	mesos := megos.NewClient(masterUrls, nil)
	return mesos.GetStateFromCluster()
}

func (s *Connector) SendCall(call *sched.Call) {
	resp, err := s.send(call)
	if err != nil {
		logrus.Errorf("send call to master got err: %s", err)
		s.emitError(utils.SeverityLow, err)
		return
	}
	if code := resp.StatusCode; code != 202 {
		logrus.Errorf("send call %+v to master, expect 202, got %d", call, code)
		s.emitError(utils.SeverityLow, "sending call response status code not 202")
	}
}

func (s *Connector) send(call *sched.Call) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}
	return s.MesosLeaderHttpClient.send(payload)
}

func (s *Connector) emitEvent(eventType sched.Event_Type, e *sched.Event) {
	s.EventChan <- &event.MesosEvent{EventType: eventType, Event: e}
}

func (s *Connector) emitError(level utils.SwanErrorSeverity, err interface{}) {
	s.ErrorChan <- utils.NewError(level, err)
}

func (s *Connector) ErrEvent() chan error {
	return s.ErrorChan
}

func (s *Connector) MesosEvent() chan *event.MesosEvent {
	return s.EventChan
}

func (s *Connector) Reregister() error {
	logrus.Infof("re-register to mesos now")

	// cancel previous stale goroutine
	if s.StreamCancelFun != nil {
		s.StreamCancelFun()
	}

	err := s.leaderDetect()
	if err != nil { // if leader detect encounter any error
		logrus.Errorf("exiting re-register due to err: %v", err)
		return err
	}

	s.StreamCtx, s.StreamCancelFun = context.WithCancel(context.Background())
	go s.subscribe(s.StreamCtx)
	return nil
}

func (s *Connector) Start(ctx context.Context) {
	err := s.leaderDetect()
	if err != nil {
		logrus.Errorf("start mesos connector got error: %v", err)
		s.emitError(utils.SeverityHigh, err) // set SeverityHigh when first start
		return
	}

	s.StreamCtx, s.StreamCancelFun = context.WithCancel(context.Background())
	go s.subscribe(s.StreamCtx)
}

func (s *Connector) leaderDetect() error {
	masters, err := getMastersFromZK(instance.MesosZkPath)
	if err != nil {
		return err
	}

	state, err := stateFromMasters(masters)
	if err != nil {
		return err
	}

	s.MesosLeaderHttpClient = NewHTTPClient(state.Leader, "/api/v1/scheduler")
	s.MesosLeader = state.Leader

	s.ClusterID = "cluster"
	if v := strings.TrimSpace(state.Cluster); v != "" {
		s.ClusterID = v
	}

	if SPECIAL_CHARACTER.MatchString(s.ClusterID) {
		logrus.Warnf(`Swan do not work with mesos cluster name(%s) with special characters "-.$*+?{}()[]|".`, s.ClusterID)
		s.ClusterID = SPECIAL_CHARACTER.ReplaceAllString(s.ClusterID, "")
	}

	return nil
}
