package mesos_connector

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/src/manager/framework/event"
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
	"github.com/Dataman-Cloud/swan/src/swancontext"

	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
	"github.com/golang/protobuf/proto"
	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

var instance *MesosConnector
var once sync.Once

type MesosConnector struct {
	// mesos framework related
	ClusterID        string
	master           string
	client           *MesosHttpClient
	lastHearBeatTime time.Time

	MesosCallChan chan *sched.Call

	// TODO make sure this chan doesn't explode
	MesosEventChan chan *event.MesosEvent
	Framework      *mesos.FrameworkInfo
}

func NewMesosConnector() *MesosConnector {
	return Instance() // call initialize method
}

func Instance() *MesosConnector {
	once.Do(
		func() {
			instance = &MesosConnector{
				MesosEventChan: make(chan *event.MesosEvent, 1024), // make this unbound in future
				MesosCallChan:  make(chan *sched.Call, 1024),
			}
		})

	return instance
}

func (s *MesosConnector) subscribe(ctx context.Context, mesosFailureChan chan error) {
	logrus.Infof("Subscribe with mesos master %s", s.master)
	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: s.Framework,
		},
	}

	if s.Framework.Id != nil {
		call.FrameworkId = &mesos.FrameworkID{
			Value: proto.String(s.Framework.Id.GetValue()),
		}
	}

	resp, err := s.Send(call)
	if err != nil {
		mesosFailureChan <- err
	}

	// http might now be the default transport in future release
	if resp.StatusCode != http.StatusOK {
		mesosFailureChan <- fmt.Errorf("Subscribe with unexpected response status: %d", resp.StatusCode)
	}

	logrus.Info(s.client.StreamID)
	go s.handleEvents(ctx, resp, mesosFailureChan)
}

func (s *MesosConnector) handleEvents(ctx context.Context, resp *http.Response, mesosFailureChan chan error) {
	defer func() {
		resp.Body.Close()
	}()

	r := NewReader(resp.Body)
	dec := json.NewDecoder(r)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("handleEvents cancelled %s", ctx.Err())
			return
		default:
			event := new(sched.Event)
			if err := dec.Decode(event); err != nil {
				logrus.Errorf("Deocde event failed: %s", err)
				mesosFailureChan <- err
			}

			switch event.GetType() {
			case sched.Event_SUBSCRIBED:
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

func CreateFrameworkInfo() *mesos.FrameworkInfo {
	fw := &mesos.FrameworkInfo{
		User:            proto.String(swancontext.Instance().Config.Scheduler.MesosFrameworkUser),
		Name:            proto.String("swan"),
		FailoverTimeout: proto.Float64(60 * 60 * 24 * 7),
	}

	return fw
}

func getMastersFromZK(zkPath string) ([]string, error) {
	masterInfo := new(mesos.MasterInfo)

	connUrl := zkPath
	if !strings.HasPrefix(connUrl, "zk://") {
		connUrl = fmt.Sprintf("zk://%s", zkPath)
	}
	url, err := url.Parse(connUrl)
	if err != nil {
		return nil, err
	}

	conn, _, err := zk.Connect(strings.Split(url.Host, ","), time.Second)
	defer conn.Close()
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to zookeeper:%s", err.Error())
	}

	// find mesos master
	children, _, err := conn.Children(url.Path)
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to zookeeper:%s", err.Error())
	}

	masters := make([]string, 0)
	for _, node := range children {
		if strings.HasPrefix(node, "json.info") {
			data, _, _ := conn.Get(url.Path + "/" + node)
			err := json.Unmarshal(data, masterInfo)
			if err != nil {
				return nil, fmt.Errorf("Unmarshal error: %s", err.Error())
			}
			masters = append(masters, fmt.Sprintf("%s:%d", *masterInfo.GetAddress().Ip, *masterInfo.GetAddress().Port))
		}
	}

	logrus.Info("Find mesos masters: ", masters)
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

func (s *MesosConnector) Send(call *sched.Call) (*http.Response, error) {
	payload, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}
	return s.client.Send(payload)
}

func (s *MesosConnector) addEvent(eventType sched.Event_Type, e *sched.Event) {
	s.MesosEventChan <- &event.MesosEvent{EventType: eventType, Event: e}
}

func (s *MesosConnector) Start(ctx context.Context, mesosFailureChan chan error) {
	var err error
	masters, err := getMastersFromZK(swancontext.Instance().Config.Scheduler.ZkPath)
	if err != nil {
		logrus.Error(err)
		return
	}
	state, err := stateFromMasters(masters)
	if err != nil {
		logrus.Errorf("%s Check your mesos master configuration", err)
		mesosFailureChan <- err
		return
	}

	s.master = state.Leader
	s.client = NewHTTPClient(state.Leader, "/api/v1/scheduler")

	s.ClusterID = state.Cluster
	if s.ClusterID == "" {
		s.ClusterID = "cluster"
	}

	r, _ := regexp.Compile("([\\-\\.\\$\\*\\+\\?\\{\\}\\(\\)\\[\\]\\|]+)")
	match := r.MatchString(s.ClusterID)
	if match {
		logrus.Warnf(`Swan do not work with mesos cluster name(%s) with special characters "-.$*+?{}()[]|".`, s.ClusterID)
		s.ClusterID = r.ReplaceAllString(s.ClusterID, "")
		logrus.Infof("Swan acceptable cluster name: %s", s.ClusterID)
	}

	s.subscribe(ctx, mesosFailureChan)

	for {
		select {
		case <-ctx.Done():
			logrus.Errorf("mesosConnector got signal %s", ctx.Err())
			return
		case call := <-s.MesosCallChan:
			logrus.WithFields(logrus.Fields{"sending-call": sched.Call_Type_name[int32(*call.Type)]}).Debugf("%+v", call)
			resp, err := s.Send(call)
			if err != nil {
				logrus.Errorf("%s", err)
				mesosFailureChan <- err
			}
			if resp.StatusCode != 202 {
				logrus.Infof("send response not 202 but %d", resp.StatusCode)
				mesosFailureChan <- errors.New("http got respose not 202")
			}
		}
	}
}
