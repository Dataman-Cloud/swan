package janitor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Dataman-Cloud/swan/agent/janitor/stats"
	"github.com/Dataman-Cloud/swan/agent/janitor/upstream"
)

func (s *JanitorServer) ListUpstreams(c *gin.Context) {
	c.JSON(200, upstream.AllUpstreams())
}

func (s *JanitorServer) GetUpstream(c *gin.Context) {
	var (
		uid = c.Param("uid")
		m   = upstream.AllUpstreams()
		ret = new(upstream.Upstream)
	)

	for idx, u := range m {
		if u.Name == uid {
			ret = m[idx]
			break
		}
	}

	c.JSON(200, ret)
}

func (s *JanitorServer) UpsertUpstream(c *gin.Context) {
	var cmb *upstream.BackendCombined
	if err := c.BindJSON(&cmb); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	if err := cmb.Valid(); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	if err := s.UpsertBackend(cmb); err != nil {
		http.Error(c.Writer, err.Error(), 500)
		return
	}

	c.Writer.WriteHeader(201)
}

func (s *JanitorServer) DelUpstream(c *gin.Context) {
	var cmb *upstream.BackendCombined
	if err := c.BindJSON(&cmb); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	s.removeBackend(cmb)
	c.Writer.WriteHeader(204)
}

func (s *JanitorServer) ListSessions(c *gin.Context) {
	c.JSON(200, upstream.AllSessions())
}

func (s *JanitorServer) ShowConfigs(c *gin.Context) {
	c.JSON(200, s.config)
}

func (s *JanitorServer) ShowStats(c *gin.Context) {
	wrapper := map[string]interface{}{
		"httpd":    s.config.ListenAddr,
		"httpdTLS": s.config.TLSListenAddr,
		"counter":  stats.Get(),
		"tcpd":     s.tcpd,
	}
	c.JSON(200, wrapper)
}

func (s *JanitorServer) ShowUpstreamStats(c *gin.Context) {
	uid := c.Param("uid")
	if m, ok := stats.UpstreamStats()[uid]; ok {
		c.JSON(200, m)
		return
	}
	c.JSON(200, make(map[string]interface{}))
}

func (s *JanitorServer) ShowBackendStats(c *gin.Context) {
	uid, bid := c.Param("uid"), c.Param("bid")
	if ups, ok := stats.UpstreamStats()[uid]; ok {
		if backend, ok := ups[bid]; ok {
			c.JSON(200, backend)
			return
		}
	}
	c.JSON(200, make(map[string]interface{}))
}
