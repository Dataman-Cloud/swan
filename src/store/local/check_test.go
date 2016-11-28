package boltdb

import (
	"os"
	"testing"

	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestSaveCheck(t *testing.T) {
	bolt, _ := NewTestBoltStore("/tmp/boltdbtest")
	defer func() {
		bolt.Close()
		os.Remove("/tmp/boltdbtest")
	}()

	task := &types.Task{
		Name:          "xxxxxx",
		AgentHostname: proto.String("x.x.x.x"),
		HealthChecks: []*types.HealthCheck{
			&types.HealthCheck{
				Protocol:        "http",
				IntervalSeconds: 2,
				TimeoutSeconds:  2,
				Command: &types.Command{
					Value: "xxxxx",
				},

				Path: proto.String("/"),
			},
		},
	}

	bolt.SaveCheck(task, 8080, "abc")

	checks, _ := bolt.ListChecks()
	assert.Equal(t, checks[0].Port, 8080)
	assert.Equal(t, checks[0].ID, "xxxxxx")

	bolt.DeleteCheck("xxxxxx")
	checks, _ = bolt.ListChecks()
	assert.Equal(t, len(checks), 0)
}
