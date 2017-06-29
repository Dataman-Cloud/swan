package nameserver

import "github.com/gin-gonic/gin"

func (s *Resolver) ApiServe(r *gin.RouterGroup) {
	r.GET("", s.listAllRecords)
	r.GET("/configs", s.showConfigs)
}

func (s *Resolver) listAllRecords(c *gin.Context) {
	c.JSON(200, s.allRecords())
}

func (s *Resolver) showConfigs(c *gin.Context) {
	c.JSON(200, s.config)
}
