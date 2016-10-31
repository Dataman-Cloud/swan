package scheduler

import (
	"github.com/Dataman-Cloud/swan/mesosproto/mesos"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPorts(t *testing.T) {
	offer := &mesos.Offer{}

	offer.Resources = append(offer.Resources, &mesos.Resource{
		Name: proto.String("ports"),
		Type: mesos.Value_RANGES.Enum(),
		Ranges: &mesos.Value_Ranges{
			Range: []*mesos.Value_Range{
				{
					Begin: proto.Uint64(uint64(1000)),
					End:   proto.Uint64(uint64(1001)),
				},
			},
		},
	})

	ports := GetPorts(offer)

	result := []uint64{1000, 1001}
	assert.Equal(t, result, ports)
}
