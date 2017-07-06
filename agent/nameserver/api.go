package nameserver

import "github.com/gin-gonic/gin"

func (s *Resolver) ListAllRecords(c *gin.Context) {
	c.JSON(200, s.allRecords())
}

func (s *Resolver) ShowConfigs(c *gin.Context) {
	c.JSON(200, s.config)
}
