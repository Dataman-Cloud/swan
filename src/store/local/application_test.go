package boltdb

import (
	"os"
	"testing"

	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/boltdb/bolt"

	"github.com/stretchr/testify/assert"
)

func NewTestBoltStore(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return NewBoltStore(db)
}

func TestSaveApplication(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app := &types.Application{
		ID:   "test",
		Name: "testapp",
	}

	bolt.SaveApplication(app)

	app, _ = bolt.FetchApplication("test")

	assert.Equal(t, app.Name, "testapp")
}

func TestListApplications(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app1 := &types.Application{
		ID:   "test1",
		Name: "testapp1",
	}

	bolt.SaveApplication(app1)

	app2 := &types.Application{
		ID:   "test2",
		Name: "testapp2",
	}

	bolt.SaveApplication(app2)

	apps, _ := bolt.ListApplications()

	assert.Equal(t, len(apps), 2)
	assert.Equal(t, apps[0].ID, "test1")
	assert.Equal(t, apps[1].Name, "testapp2")
}

func TestDeleteApplication(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	app := &types.Application{
		ID:   "test",
		Name: "testapp",
	}

	bolt.SaveApplication(app)

	apps, _ := bolt.ListApplications()

	assert.Equal(t, len(apps), 1)

	bolt.DeleteApplication("test")

	apps, _ = bolt.ListApplications()

	assert.Equal(t, len(apps), 0)
}

func TestIncreaseApplicationUpdatedInstances(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()
	app := &types.Application{
		ID:               "test",
		Name:             "testapp",
		UpdatedInstances: 0,
	}

	bolt.SaveApplication(app)

	bolt.IncreaseApplicationUpdatedInstances("test")

	app, _ = bolt.FetchApplication("test")

	assert.Equal(t, int(app.UpdatedInstances), 1)
}

func TestIncreaseApplicationInstances(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()
	app := &types.Application{
		ID:        "test",
		Name:      "testapp",
		Instances: 0,
	}

	bolt.SaveApplication(app)
	bolt.IncreaseApplicationInstances("test")
	app, _ = bolt.FetchApplication("test")
	assert.Equal(t, int(app.Instances), 1)

}

func TestResetApplicationUpdatedInstances(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()
	app := &types.Application{
		ID:               "test",
		Name:             "testapp",
		UpdatedInstances: 1,
	}

	bolt.SaveApplication(app)
	bolt.ResetApplicationUpdatedInstances("test")

	app, _ = bolt.FetchApplication("test")

	assert.Equal(t, int(app.UpdatedInstances), 0)
}

func TestUpdateApplicationStatus(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()
	app := &types.Application{
		ID:     "test",
		Name:   "testapp",
		Status: "STAGING",
	}

	bolt.SaveApplication(app)
	bolt.UpdateApplicationStatus("test", "RUNNING")

	app, _ = bolt.FetchApplication("test")

	assert.Equal(t, app.Status, "RUNNING")
}

func TestIncreaseApplicationRunningInstances(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()
	app := &types.Application{
		ID:               "test",
		Name:             "testapp",
		RunningInstances: 0,
	}

	bolt.SaveApplication(app)
	bolt.IncreaseApplicationRunningInstances("test")

	app, _ = bolt.FetchApplication("test")

	assert.Equal(t, int(app.RunningInstances), 1)
}

func TestReduceApplicationRunningInstances(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()
	app := &types.Application{
		ID:               "test",
		Name:             "testapp",
		RunningInstances: 1,
	}

	bolt.SaveApplication(app)
	bolt.ReduceApplicationRunningInstances("test")

	app, _ = bolt.FetchApplication("test")

	assert.Equal(t, int(app.RunningInstances), 0)
}

func TestReduceApplicationInstances(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()
	app := &types.Application{
		ID:        "test",
		Name:      "testapp",
		Instances: 1,
	}

	bolt.SaveApplication(app)
	bolt.ReduceApplicationInstances("test")
	app, _ = bolt.FetchApplication("test")
	assert.Equal(t, int(app.Instances), 0)
}
