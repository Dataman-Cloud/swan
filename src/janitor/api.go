package janitor

import "github.com/gin-gonic/gin"

func (s *JanitorServer) ApiServe(r *gin.RouterGroup) {
	r.GET("", s.listUpstreams)
	r.GET("/upstreams", s.listUpstreams)
	r.GET("/sessions", s.listSessions)
	r.GET("/configs", s.showConfigs)
	r.GET("/stats", s.showStats)
	r.GET("/stats/:aid", s.showAppStats)
	r.GET("/stats/:aid/:tid", s.showTaskStats)
}

func (s *JanitorServer) listUpstreams(c *gin.Context) {
	c.JSON(200, s.upstreams.allUps())
}

func (s *JanitorServer) listSessions(c *gin.Context) {
	c.JSON(200, s.upstreams.allSess())
}

func (s *JanitorServer) showConfigs(c *gin.Context) {
	c.JSON(200, s.config)
}

func (s *JanitorServer) showStats(c *gin.Context) {
	c.JSON(200, s.stats)
}

func (s *JanitorServer) showAppStats(c *gin.Context) {
	aid := c.Param("aid")
	if m, ok := s.stats.App[aid]; ok {
		c.JSON(200, m)
		return
	}
	c.JSON(200, make(map[string]interface{}))
}

func (s *JanitorServer) showTaskStats(c *gin.Context) {
	aid, tid := c.Param("aid"), c.Param("tid")
	if a, ok := s.stats.App[aid]; ok {
		if t, ok := a[tid]; ok {
			c.JSON(200, t)
			return
		}
	}
	c.JSON(200, make(map[string]interface{}))
}
