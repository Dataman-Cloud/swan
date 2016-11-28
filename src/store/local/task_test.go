package boltdb

import (
	"github.com/Dataman-Cloud/swan/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSaveTask(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	task := &types.Task{
		ID:   "xxxyyy",
		Name: "x.y.z",
	}

	bolt.SaveTask(task)

	task, _ = bolt.FetchTask("x.y.z")

	assert.Equal(t, task.Name, "x.y.z")
}

func TestFetchTask(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	task, _ := bolt.FetchTask("x.y.z")
	assert.Nil(t, task)
}

func TestListTasks(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	task1 := &types.Task{
		ID:    "xxxyyy",
		Name:  "x.y.z",
		AppId: "m",
	}

	bolt.SaveTask(task1)

	task2 := &types.Task{
		ID:    "mmmnnn",
		Name:  "m.n.l",
		AppId: "m",
	}

	bolt.SaveTask(task2)

	tasks, _ := bolt.ListTasks("m")
	assert.Equal(t, len(tasks), 2)
}

func TestDeleteTask(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	task1 := &types.Task{
		ID:    "xxxyyy",
		Name:  "x.y.z",
		AppId: "m",
	}

	bolt.SaveTask(task1)

	bolt.DeleteTask("xxxyyy")

	task, _ := bolt.FetchTask("xxxyyy")
	assert.Nil(t, task)
}

func TestUpdateTaskStatus(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	task1 := &types.Task{
		ID:     "xxxyyy",
		Name:   "x.y.z",
		Status: "RUNNING",
	}

	bolt.SaveTask(task1)

	bolt.UpdateTaskStatus("x.y.z", "FAILED")

	task, _ := bolt.FetchTask("x.y.z")
	assert.Equal(t, task.Status, "FAILED")

	err := bolt.UpdateTaskStatus("x.y.z.m.n", "RUNNING")
	assert.NotNil(t, err)
}

func TestDeleteApplicationTasks(t *testing.T) {
	bolt, _ := NewBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	task := &types.Task{
		ID:     "xxxyyy",
		Name:   "x.y.z",
		AppId:  "testapp",
		Status: "RUNNING",
	}

	bolt.SaveTask(task)

	bolt.DeleteApplicationTasks("testapp")

	tasks, _ := bolt.ListTasks("testapp")

	assert.Equal(t, len(tasks), 0)
}
