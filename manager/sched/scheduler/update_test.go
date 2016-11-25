package scheduler

import (
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/mesosproto/sched"
	. "github.com/Dataman-Cloud/swan/store/local"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestStatusRunning(t *testing.T) {
	f := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/scheduler", f)
	srv := httptest.NewServer(m)
	defer srv.Close()

	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app := &types.Application{
		ID:               "bb",
		Name:             "bb",
		RunningInstances: 0,
		Instances:        1,
	}

	bolt.SaveApplication(app)

	task := &types.Task{
		ID:     "aa.bb.cc.dd",
		Name:   "aa.bb.cc.dd",
		AppId:  "bb",
		Status: "",
	}

	bolt.SaveTask(task)

	s := NewScheduler(strings.TrimPrefix(srv.URL, "http://"), nil, bolt, "xxxxx", nil, nil, nil)
	ev := &sched.Event{
		Type: sched.Event_UPDATE.Enum(),
		Update: &sched.Event_Update{
			Status: &mesos.TaskStatus{
				TaskId: &mesos.TaskID{
					Value: proto.String("xxxxxx-aa.bb.cc.dd"),
				},
				State:   mesos.TaskState_TASK_RUNNING.Enum(),
				Message: proto.String("zzzzzzz"),
				Uuid:    []byte(`xxxxxxxxxx`),
				AgentId: &mesos.AgentID{
					Value: proto.String("yyyyyyyyy"),
				},
			},
		},
	}
	status := ev.GetUpdate().GetStatus()
	s.status(status)
}

func TestStatusSTAGING(t *testing.T) {
	f := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/scheduler", f)
	srv := httptest.NewServer(m)
	defer srv.Close()

	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app := &types.Application{
		ID:               "bb",
		Name:             "bb",
		RunningInstances: 0,
		Instances:        1,
	}

	bolt.SaveApplication(app)

	task := &types.Task{
		ID:     "aa.bb.cc.dd",
		Name:   "aa.bb.cc.dd",
		AppId:  "bb",
		Status: "",
	}

	bolt.SaveTask(task)

	s := NewScheduler(strings.TrimPrefix(srv.URL, "http://"), nil, bolt, "xxxxx", nil, nil, nil)
	ev := &sched.Event{
		Type: sched.Event_UPDATE.Enum(),
		Update: &sched.Event_Update{
			Status: &mesos.TaskStatus{
				TaskId: &mesos.TaskID{
					Value: proto.String("xxxxxx-aa.bb.cc.dd"),
				},
				State:   mesos.TaskState_TASK_STAGING.Enum(),
				Message: proto.String("zzzzzzz"),
				Uuid:    []byte(`xxxxxxxxxx`),
				AgentId: &mesos.AgentID{
					Value: proto.String("yyyyyyyyy"),
				},
			},
		},
	}
	status := ev.GetUpdate().GetStatus()
	s.status(status)
}

func TestStatusSTARTING(t *testing.T) {
	f := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/scheduler", f)
	srv := httptest.NewServer(m)
	defer srv.Close()

	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app := &types.Application{
		ID:               "bb",
		Name:             "bb",
		RunningInstances: 0,
		Instances:        1,
	}

	bolt.SaveApplication(app)

	task := &types.Task{
		ID:     "aa.bb.cc.dd",
		Name:   "aa.bb.cc.dd",
		AppId:  "bb",
		Status: "",
	}

	bolt.SaveTask(task)

	s := NewScheduler(strings.TrimPrefix(srv.URL, "http://"), nil, bolt, "xxxxx", nil, nil, nil)
	ev := &sched.Event{
		Type: sched.Event_UPDATE.Enum(),
		Update: &sched.Event_Update{
			Status: &mesos.TaskStatus{
				TaskId: &mesos.TaskID{
					Value: proto.String("xxxxxx-aa.bb.cc.dd"),
				},
				State:   mesos.TaskState_TASK_STARTING.Enum(),
				Message: proto.String("zzzzzzz"),
				Uuid:    []byte(`xxxxxxxxxx`),
				AgentId: &mesos.AgentID{
					Value: proto.String("yyyyyyyyy"),
				},
			},
		},
	}
	status := ev.GetUpdate().GetStatus()
	s.status(status)
}

func TestStatusFINISHED(t *testing.T) {
	f := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/scheduler", f)
	srv := httptest.NewServer(m)
	defer srv.Close()

	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app := &types.Application{
		ID:               "bb",
		Name:             "bb",
		RunningInstances: 0,
		Instances:        1,
	}

	bolt.SaveApplication(app)

	task := &types.Task{
		ID:     "aa.bb.cc.dd",
		Name:   "aa.bb.cc.dd",
		AppId:  "bb",
		Status: "",
	}

	bolt.SaveTask(task)

	s := NewScheduler(strings.TrimPrefix(srv.URL, "http://"), nil, bolt, "xxxxx", nil, nil, nil)
	ev := &sched.Event{
		Type: sched.Event_UPDATE.Enum(),
		Update: &sched.Event_Update{
			Status: &mesos.TaskStatus{
				TaskId: &mesos.TaskID{
					Value: proto.String("xxxxxx-aa.bb.cc.dd"),
				},
				State:   mesos.TaskState_TASK_FINISHED.Enum(),
				Message: proto.String("zzzzzzz"),
				Uuid:    []byte(`xxxxxxxxxx`),
				AgentId: &mesos.AgentID{
					Value: proto.String("yyyyyyyyy"),
				},
			},
		},
	}
	status := ev.GetUpdate().GetStatus()
	s.status(status)
}

func TestStatusFAILED(t *testing.T) {
	f := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/scheduler", f)
	srv := httptest.NewServer(m)
	defer srv.Close()

	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app := &types.Application{
		ID:               "bb",
		Name:             "bb",
		RunningInstances: 0,
		Instances:        1,
	}

	bolt.SaveApplication(app)

	task := &types.Task{
		ID:     "aa.bb.cc.dd",
		Name:   "aa.bb.cc.dd",
		AppId:  "bb",
		Status: "",
	}

	bolt.SaveTask(task)

	s := NewScheduler(strings.TrimPrefix(srv.URL, "http://"), nil, bolt, "xxxxx", nil, nil, nil)
	ev := &sched.Event{
		Type: sched.Event_UPDATE.Enum(),
		Update: &sched.Event_Update{
			Status: &mesos.TaskStatus{
				TaskId: &mesos.TaskID{
					Value: proto.String("xxxxxx-aa.bb.cc.dd"),
				},
				State:   mesos.TaskState_TASK_FAILED.Enum(),
				Message: proto.String("zzzzzzz"),
				Uuid:    []byte(`xxxxxxxxxx`),
				AgentId: &mesos.AgentID{
					Value: proto.String("yyyyyyyyy"),
				},
			},
		},
	}
	status := ev.GetUpdate().GetStatus()
	s.status(status)
}

func TestStatusKILLED(t *testing.T) {
	f := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/scheduler", f)
	srv := httptest.NewServer(m)
	defer srv.Close()

	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app := &types.Application{
		ID:               "bb",
		Name:             "bb",
		RunningInstances: 0,
		Instances:        1,
	}

	bolt.SaveApplication(app)

	task := &types.Task{
		ID:     "aa.bb.cc.dd",
		Name:   "aa.bb.cc.dd",
		AppId:  "bb",
		Status: "",
	}

	bolt.SaveTask(task)

	s := NewScheduler(strings.TrimPrefix(srv.URL, "http://"), nil, bolt, "xxxxx", nil, nil, nil)
	ev := &sched.Event{
		Type: sched.Event_UPDATE.Enum(),
		Update: &sched.Event_Update{
			Status: &mesos.TaskStatus{
				TaskId: &mesos.TaskID{
					Value: proto.String("xxxxxx-aa.bb.cc.dd"),
				},
				State:   mesos.TaskState_TASK_KILLED.Enum(),
				Message: proto.String("zzzzzzz"),
				Uuid:    []byte(`xxxxxxxxxx`),
				AgentId: &mesos.AgentID{
					Value: proto.String("yyyyyyyyy"),
				},
			},
		},
	}
	status := ev.GetUpdate().GetStatus()
	s.status(status)
}

func TestStatusLOST(t *testing.T) {
	f := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
	m := mux.NewRouter()
	m.HandleFunc("/api/v1/scheduler", f)
	srv := httptest.NewServer(m)
	defer srv.Close()

	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app := &types.Application{
		ID:               "bb",
		Name:             "bb",
		RunningInstances: 0,
		Instances:        1,
	}

	bolt.SaveApplication(app)

	task := &types.Task{
		ID:     "aa.bb.cc.dd",
		Name:   "aa.bb.cc.dd",
		AppId:  "bb",
		Status: "",
	}

	bolt.SaveTask(task)

	s := NewScheduler(strings.TrimPrefix(srv.URL, "http://"), nil, bolt, "xxxxx", nil, nil, nil)
	ev := &sched.Event{
		Type: sched.Event_UPDATE.Enum(),
		Update: &sched.Event_Update{
			Status: &mesos.TaskStatus{
				TaskId: &mesos.TaskID{
					Value: proto.String("xxxxxx-aa.bb.cc.dd"),
				},
				State:   mesos.TaskState_TASK_LOST.Enum(),
				Message: proto.String("zzzzzzz"),
				Uuid:    []byte(`xxxxxxxxxx`),
				AgentId: &mesos.AgentID{
					Value: proto.String("yyyyyyyyy"),
				},
			},
		},
	}
	status := ev.GetUpdate().GetStatus()
	s.status(status)
}
