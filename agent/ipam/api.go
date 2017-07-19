package ipam

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *IPAM) ListSubNets(c *gin.Context) {
	subnets, err := m.store.ListSubNets()
	if err != nil {
		http.Error(c.Writer, err.Error(), 500)
		return
	}

	// wrap each subnet with ips
	ret := make(map[string]interface{})
	for _, subnet := range subnets {
		ips, err := m.store.ListIPs(subnet.ID)
		if err != nil {
			ret[subnet.ID] = map[string]interface{}{
				"subnet": subnet,
				"ips":    nil,
				"err":    err.Error(),
			}
		} else {
			ret[subnet.ID] = map[string]interface{}{
				"subnet": subnet,
				"ips":    ips,
			}
		}
	}

	c.JSON(200, ret)
}

func (m *IPAM) SetSubNetPool(c *gin.Context) {
}
