package janitor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Dataman-Cloud/swan/src/janitor/stats"
	"github.com/Dataman-Cloud/swan/src/janitor/upstream"
)

func (s *JanitorServer) ApiServe(r *gin.RouterGroup) {
	r.GET("", s.listUpstreams)
	r.GET("/upstreams", s.listUpstreams)
	r.PUT("/upstreams", s.upsertUpstream)
	r.DELETE("/upstreams", s.delUpstream)
	r.GET("/sessions", s.listSessions)
	r.GET("/configs", s.showConfigs)
	r.GET("/stats", s.showStats)
	r.GET("/stats/:uid", s.showUpstreamStats)
	r.GET("/stats/:uid/:bid", s.showBackendStats)
}

func (s *JanitorServer) listUpstreams(c *gin.Context) {
	c.JSON(200, upstream.AllUpstreams())
}

func (s *JanitorServer) upsertUpstream(c *gin.Context) {
	var cmb *upstream.BackendCombined
	if err := c.BindJSON(&cmb); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	if err := cmb.Valid(); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	if err := s.upsertBackend(cmb); err != nil {
		http.Error(c.Writer, err.Error(), 500)
		return
	}

	c.Writer.WriteHeader(201)
}

func (s *JanitorServer) delUpstream(c *gin.Context) {
	var cmb *upstream.BackendCombined
	if err := c.BindJSON(&cmb); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	s.removeBackend(cmb)
	c.Writer.WriteHeader(204)
}

func (s *JanitorServer) listSessions(c *gin.Context) {
	c.JSON(200, upstream.AllSessions())
}

func (s *JanitorServer) showConfigs(c *gin.Context) {
	c.JSON(200, s.config)
}

func (s *JanitorServer) showStats(c *gin.Context) {
	wrapper := map[string]interface{}{
		"httpd":          s.config.ListenAddr,
		"httpdTLS":       s.config.TLSListenAddr,
		"queuing_events": len(s.eventChan),
		"counter":        stats.Get(),
		"tcpd":           s.tcpd,
	}
	c.JSON(200, wrapper)
}

func (s *JanitorServer) showUpstreamStats(c *gin.Context) {
	uid := c.Param("uid")
	if m, ok := stats.UpstreamStats()[uid]; ok {
		c.JSON(200, m)
		return
	}
	c.JSON(200, make(map[string]interface{}))
}

func (s *JanitorServer) showBackendStats(c *gin.Context) {
	uid, bid := c.Param("uid"), c.Param("bid")
	if ups, ok := stats.UpstreamStats()[uid]; ok {
		if backend, ok := ups[bid]; ok {
			c.JSON(200, backend)
			return
		}
	}
	c.JSON(200, make(map[string]interface{}))
}
