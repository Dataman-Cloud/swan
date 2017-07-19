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

	// /proxy/**
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

	// /dns/**
	r = mux.Group("/dns")
	r.GET("", resolver.ListRecords)
	r.GET("/records", resolver.ListRecords)
	r.PUT("/records", resolver.UpsertRecord)
	r.DELETE("/records", resolver.DelRecord)
	r.GET("/configs", resolver.ShowConfigs)
	r.GET("/stats", resolver.ShowStats)

	// /ipam/**
	r = mux.Group("/ipam")
	r.GET("", ipam.ListSubNets)
	r.GET("subnets", ipam.ListSubNets)

	mux.NoRoute(agent.serveProxy)
	return mux
}
