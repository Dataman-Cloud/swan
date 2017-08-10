package resolver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Resolver) ListRecords(c *gin.Context) {
	c.JSON(200, s.allRecords())
}

func (s *Resolver) GetRecord(c *gin.Context) {
	var (
		id  = c.Param("id")
		m   = s.allRecords()
		ret = make([]*Record, 0)
	)
	if val, ok := m[id]; ok {
		ret = val
	}
	c.JSON(200, ret)
}

func (s *Resolver) UpsertRecord(c *gin.Context) {
	var record *Record
	if err := c.BindJSON(&record); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	if err := s.Upsert(record); err != nil {
		http.Error(c.Writer, err.Error(), 500)
		return
	}

	c.Writer.WriteHeader(201)

}

func (s *Resolver) DelRecord(c *gin.Context) {
	var record *Record
	if err := c.BindJSON(&record); err != nil {
		http.Error(c.Writer, err.Error(), 400)
		return
	}

	s.remove(record)
	c.Writer.WriteHeader(204)
}

func (s *Resolver) ShowConfigs(c *gin.Context) {
	c.JSON(200, s.config)
}

func (s *Resolver) ShowStats(c *gin.Context) {
}
