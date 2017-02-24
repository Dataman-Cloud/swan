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

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/swancontext"
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

	MesosCallChan  chan *sched.Call
	MesosEventChan chan *event.MesosEvent

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
				User:      proto.String(swancontext.Instance().Config.Scheduler.MesosFrameworkUser),
				Name:      proto.String("swan"),
				Principal: proto.String("swan"),

				FailoverTimeout: proto.Float64(60 * 60 * 24 * 7),
				Checkpoint:      proto.Bool(true),
				Hostname:        proto.String(hostname),
			}

			instance = &Connector{
				MesosEventChan: make(chan *event.MesosEvent, 1024), // make this unbound in future
				MesosCallChan:  make(chan *sched.Call, 1024),
				FrameworkInfo:  info,
			}
		})

	return instance
}

func (s *Connector) Subscribe(ctx context.Context, mesosFailureChan chan error) {
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
		mesosFailureChan <- err
		return // shortcut this without further actions
	}

	if resp.StatusCode != http.StatusOK {
		mesosFailureChan <- fmt.Errorf("subscribe with unexpected response status: %d", resp.StatusCode)
		return // shortcut this without further actions
	}

	s.handleEvents(ctx, resp, mesosFailureChan)
}

func (s *Connector) handleEvents(ctx context.Context, resp *http.Response, mesosFailureChan chan error) {
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
				mesosFailureChan <- err
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

	conn, _, err := zk.Connect(strings.Split(url.Host, ","), time.Second)
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
	s.MesosEventChan <- &event.MesosEvent{EventType: eventType, Event: e}
}

func (s *Connector) Reregister(mesosFailureChan chan error) {
	logrus.Infof("register to mesos now")

	if s.StreamCancelFun != nil {
		s.StreamCancelFun()
	}

	err := s.LeaderDetect()
	if err != nil { // if leader detect encounter any error
		mesosFailureChan <- utils.NewError(utils.SeverityLow, err)
		return
	}

	s.StreamCtx, s.StreamCancelFun = context.WithCancel(context.Background())
	go s.Subscribe(s.StreamCtx, mesosFailureChan)
}

func (s *Connector) Start(ctx context.Context, mesosFailureChan chan error) {
	err := s.LeaderDetect()
	if err != nil {
		mesosFailureChan <- utils.NewError(utils.SeverityHigh, err) // set SeverityHigh when first start
		return
	}

	s.StreamCtx, s.StreamCancelFun = context.WithCancel(context.Background())
	go s.Subscribe(s.StreamCtx, mesosFailureChan)

	for {
		select {
		case <-ctx.Done():
			s.StreamCancelFun() // stop stream goroutine
			logrus.Infof("connector got done signal: %s", ctx.Err())
			return
		case call := <-s.MesosCallChan:
			logrus.WithFields(logrus.Fields{"sending-call": sched.Call_Type_name[int32(*call.Type)]}).Debugf("%+v", call)
			resp, err := s.Send(call)
			if err != nil {
				mesosFailureChan <- err
			}
			if resp.StatusCode != 202 {
				mesosFailureChan <- errors.New("sending call response status code not 202")
			}
		}
	}
}

func (s *Connector) LeaderDetect() error {
	masters, err := getMastersFromZK(swancontext.Instance().Config.Scheduler.ZkPath)
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
