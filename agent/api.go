package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (agent *Agent) NewHTTPMux() http.Handler {
	var (
		janitor  = agent.janitor
		resolver = agent.resolver
		ipam     = agent.ipam
	)

	mux := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	mux.GET("/sysinfo", agent.sysinfo)
	mux.GET("/configs", agent.showConfigs)

	// /proxy/**
	if agent.config.Janitor.Enabled {
		r := mux.Group("/proxy")
		r.GET("", janitor.ListUpstreams)
		r.GET("/upstreams", janitor.ListUpstreams)
		r.PUT("/upstreams", janitor.UpsertUpstream)
		r.DELETE("/upstreams", janitor.DelUpstream)
		r.GET("/sessions", janitor.ListSessions)
		r.GET("/configs", janitor.ShowConfigs)
		r.GET("/stats", janitor.ShowStats)
		r.GET("/stats/:uid", janitor.ShowUpstreamStats)
		r.GET("/stats/:uid/:bid", janitor.ShowBackendStats)
	}

	// /dns/**
	if agent.config.DNS.Enabled {
		r := mux.Group("/dns")
		r.GET("", resolver.ListRecords)
		r.GET("/records", resolver.ListRecords)
		r.PUT("/records", resolver.UpsertRecord)
		r.DELETE("/records", resolver.DelRecord)
		r.GET("/configs", resolver.ShowConfigs)
		r.GET("/stats", resolver.ShowStats)
	}

	// /ipam/**
	if agent.config.IPAM.Enabled {
		r := mux.Group("/ipam")
		r.GET("", ipam.ListSubNets)
		r.GET("subnets", ipam.ListSubNets)
		r.PUT("subnets", ipam.SetSubNetPool)
	}

	mux.NoRoute(agent.serveProxy)
	return mux
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
