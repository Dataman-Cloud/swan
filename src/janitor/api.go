package janitor

import "github.com/gin-gonic/gin"

func (s *JanitorServer) ApiServe(r *gin.RouterGroup) {
	r.GET("", s.listUpstreams)
	r.GET("/upstreams", s.listUpstreams)
	r.GET("/sessions", s.listSessions)
	r.GET("/stats", s.showStats)
	r.GET("/configs", s.showConfigs)
}

func (s *JanitorServer) listUpstreams(c *gin.Context) {
	c.JSON(200, s.upstreams.allUps())
}

func (s *JanitorServer) listSessions(c *gin.Context) {
	c.JSON(200, s.upstreams.allSess())
}

func (s *JanitorServer) showStats(c *gin.Context) {
	c.JSON(200, s.stats)
}

func (s *JanitorServer) showConfigs(c *gin.Context) {
	c.JSON(200, s.config)
}
