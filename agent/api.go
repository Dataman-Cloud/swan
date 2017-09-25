package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (agent *Agent) NewHTTPMux() http.Handler {
	mux := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	mux.GET("/sysinfo", agent.sysinfo)
	mux.GET("/configs", agent.showConfigs)

	// /proxy/**
	if agent.config.Janitor.Enabled {
		agent.setupProxyHandlers(mux)
	}

	// /dns/**
	if agent.config.DNS.Enabled {
		agent.setupDNSHandlers(mux)
	}

	// /ipam/**
	if agent.config.IPAM.Enabled {
		agent.setupIPAMHandlers(mux)
	}

	mux.NoRoute(agent.serveProxy)
	return mux
}

func (agent *Agent) setupProxyHandlers(mux *gin.Engine) {
	var (
		janitor = agent.janitor
	)

	r := mux.Group("/proxy")
	r.GET("", janitor.ListUpstreams)
	r.GET("/upstreams", janitor.ListUpstreams)
	r.GET("/upstreams/:uid", janitor.GetUpstream)
	r.PUT("/upstreams", janitor.UpsertUpstream)
	r.DELETE("/upstreams", janitor.DelUpstream)
	r.GET("/sessions", janitor.ListSessions)
	r.GET("/configs", janitor.ShowConfigs)
	r.GET("/stats", janitor.ShowStats)
	r.GET("/stats/:uid", janitor.ShowUpstreamStats)
	r.GET("/stats/:uid/:bid", janitor.ShowBackendStats)
}

func (agent *Agent) setupDNSHandlers(mux *gin.Engine) {
	var (
		resolver = agent.resolver
	)

	r := mux.Group("/dns")
	r.GET("", resolver.ListRecords)
	r.GET("/records", resolver.ListRecords)
	r.GET("/records/:id", resolver.GetRecord)
	r.PUT("/records", resolver.UpsertRecord)
	r.DELETE("/records", resolver.DelRecord)
	r.GET("/configs", resolver.ShowConfigs)
	r.GET("/stats", resolver.ShowStats)
	r.GET("/stats/:id", resolver.ShowParentStats)

}

func (agent *Agent) setupIPAMHandlers(mux *gin.Engine) {
	var (
		ipam = agent.ipam
	)

	r := mux.Group("/ipam")
	r.GET("", ipam.ListSubNets)
	r.GET("subnets", ipam.ListSubNets)
	r.PUT("subnets", ipam.SetSubNetPool)
}

func (agent *Agent) sysinfo(ctx *gin.Context) {
	info, err := Gather()
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx.JSON(200, info)
}

func (agent *Agent) showConfigs(ctx *gin.Context) {
	ctx.JSON(200, agent.config)
}
