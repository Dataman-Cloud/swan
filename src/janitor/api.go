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
	r.GET("/stats/:aid", s.showAppStats)
	r.GET("/stats/:aid/:tid", s.showTaskStats)
}

func (s *JanitorServer) listUpstreams(c *gin.Context) {
	c.JSON(200, upstream.AllUpstreams())
}

func (s *JanitorServer) upsertUpstream(c *gin.Context) {
	var target *upstream.Target
	if err := c.BindJSON(&target); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	if err := target.Valid(); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	if err := s.upsertBackend(target); err != nil {
		http.Error(c.Writer, err.Error(), 500)
		return
	}

	c.Writer.WriteHeader(201)
}

func (s *JanitorServer) delUpstream(c *gin.Context) {
	var target *upstream.Target
	if err := c.BindJSON(&target); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	s.removeBackend(target)
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

func (s *JanitorServer) showAppStats(c *gin.Context) {
	aid := c.Param("aid")
	if m, ok := stats.AppStats()[aid]; ok {
		c.JSON(200, m)
		return
	}
	c.JSON(200, make(map[string]interface{}))
}

func (s *JanitorServer) showTaskStats(c *gin.Context) {
	aid, tid := c.Param("aid"), c.Param("tid")
	if a, ok := stats.AppStats()[aid]; ok {
		if t, ok := a[tid]; ok {
			c.JSON(200, t)
			return
		}
	}
	c.JSON(200, make(map[string]interface{}))
}
