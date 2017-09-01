package api

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (s *Server) setupRoutes(mux *mux.Router) {
	routes := []*Route{
		NewRoute("GET", "/v1/apps", s.listApps),
		NewRoute("POST", "/v1/apps", s.createApp),
		NewRoute("GET", "/v1/apps/{app_id}", s.getApp),
		NewRoute("DELETE", "/v1/apps/{app_id}", s.deleteApp),
		NewRoute("POST", "/v1/apps/{app_id}/scale", s.scaleApp),
		NewRoute("PUT", "/v1/apps/{app_id}", s.updateApp),
		NewRoute("POST", "/v1/apps/{app_id}/start", s.startApp),
		NewRoute("POST", "/v1/apps/{app_id}/stop", s.stopApp),
		NewRoute("PUT", "/v1/apps/{app_id}/canary", s.canaryUpdate),
		NewRoute("POST", "/v1/apps/{app_id}/rollback", s.rollback),
		NewRoute("PUT", "/v1/apps/{app_id}/weights", s.updateWeights),
		NewRoute("POST", "/v1/apps/{app_id}/reset", s.resetStatus),

		NewRoute("GET", "/v1/apps/{app_id}/tasks", s.getTasks),
		NewRoute("GET", "/v1/apps/{app_id}/tasks/{task_id}", s.getTask),
		NewRoute("DELETE", "/v1/apps/{app_id}/tasks/{task_id}", s.deleteTask),
		NewRoute("DELETE", "/v1/apps/{app_id}/tasks", s.deleteTasks),
		NewRoute("PUT", "/v1/apps/{app_id}/tasks/{task_id}", s.updateTask),
		NewRoute("POST", "/v1/apps/{app_id}/tasks/{task_id}", s.rollbackTask),
		NewRoute("PUT", "/v1/apps/{app_id}/tasks/{task_id}/weight", s.updateWeight),

		NewRoute("GET", "/v1/apps/{app_id}/versions", s.getVersions),
		NewRoute("GET", "/v1/apps/{app_id}/versions/{version_id}", s.getVersion),
		NewRoute("POST", "/v1/apps/{app_id}/versions", s.createVersion),

		// Deprecated, Remove Later
		NewRoute("POST", "/v1/compose", s.runCompose),
		NewRoute("POST", "/v1/compose/parse", s.parseYAML),
		NewRoute("GET", "/v1/compose", s.listComposes),
		NewRoute("GET", "/v1/compose/{compose_id}", s.getCompose),
		NewRoute("GET", "/v1/compose/{compose_id}/dependency", s.getComposeDependency),
		NewRoute("DELETE", "/v1/compose/{compose_id}", s.deleteCompose),

		NewRoute("POST", "/v1/compose-ng", s.runComposeNG),
		NewRoute("POST", "/v1/compose-ng/parse", s.parseYAMLNG),
		NewRoute("GET", "/v1/compose-ng", s.listComposesNG),
		NewRoute("GET", "/v1/compose-ng/{compose_id}", s.getComposeNG),
		NewRoute("GET", "/v1/compose-ng/{compose_id}/dependency", s.getComposeDependencyNG),
		NewRoute("GET", "/v1/compose-ng/{compose_id}/debug/versions", s.parseComposeToVersions),
		NewRoute("DELETE", "/v1/compose-ng/{compose_id}", s.deleteComposeNG),

		// mesos virtual cluster
		NewRoute("GET", "/v1/vclusters", s.listVClusters),
		NewRoute("POST", "/v1/vclusters", s.createVCluster),
		NewRoute("GET", "/v1/vclusters/{vcluster_name}", s.getVCluster),
		NewRoute("DELETE", "/v1/vclusters/{vcluster_name}", s.deleteVCluster),
		NewRoute("POST", "/v1/vclusters/{vcluster_name}/nodes", s.addNode),
		NewRoute("PATCH", "/v1/vclusters/{vcluster_name}/nodes/{node_ip}", s.updateNode),
		NewRoute("GET", "/v1/vclusters/{vcluster_name}/nodes", s.listNodes),
		NewRoute("GET", "/v1/vclusters/{vcluster_name}/nodes/{node_ip}", s.getNode),
		NewRoute("DELETE", "/v1/vclusters/{vcluster_name}/nodes/{node_ip}", s.delNode),

		NewRoute("GET", "/v1/mesos/agents", s.listMesosAgents),
		NewRoute("PATCH", "/v1/mesos/agents/{agent_id}", s.updateMesosAgent),

		NewRoute("GET", "/ping", s.ping),
		NewRoute("GET", "/v1/events", s.events),
		NewRoute("GET", "/v1/stats", s.stats),
		NewRoute("GET", "/version", s.version),
		NewRoute("GET", "/v1/leader", s.getLeader),
		NewRoute("POST", "/v1/purge", s.purge),

		NewRoute("GET", "/v1/debug/dump", s.dump),
		NewRoute("GET", "/v1/debug/load", s.load),
		NewRoute("GET", "/v1/fullsync", s.fullEventsAndRecords),

		NewRoute("PUT", "/v1/debug", s.enableDebug),
		NewRoute("DELETE", "/v1/debug", s.disableDebug),

		NewRoute("GET", "/v1/agents", s.listAgents),
		NewRoute("GET", "/v1/agents/{agent_id}", s.getAgent),
		NewRoute("DELETE", "/v1/agents/{agent_id}", s.closeAgent),
		NewRoute("GET", "/v1/agents/{agent_id}/sysinfo", s.getAgent),
		NewRoute("GET", "/v1/agents/{agent_id}/configs", s.getAgentConfigs),
		NewPrefixRoute("ANY", "/v1/agents/{agent_id}/proxy", s.redirectAgentProxy),
		NewPrefixRoute("ANY", "/v1/agents/{agent_id}/dns", s.redirectAgentDNS),
		NewPrefixRoute("ANY", "/v1/agents/{agent_id}/docker", s.redirectAgentDocker),
		NewPrefixRoute("ANY", "/v1/agents/{agent_id}/ipam", s.redirectAgentIPAM),

		NewRoute("GET", "/v1/apps/{app_id}/dns", s.getAppDNS),
		NewRoute("GET", "/v1/apps/{app_id}/dns/traffics", s.getAppDNSTraffics),
		NewRoute("GET", "/v1/apps/{app_id}/proxy", s.getAppProxy),
		NewRoute("GET", "/v1/apps/{app_id}/proxy/traffics", s.getAppTraffics),
	}

	log.Debug("Registering HTTP route")

	for _, r := range routes {
		var (
			handler = s.makeHTTPHandler(r.Handler())
			path    = r.Path()
			methods = r.Methods()
		)

		log.Debugf("Registering %v, %s", methods, path)

		if r.prefix {
			mux.PathPrefix(path).Methods(methods...).Handler(handler)
		} else {
			mux.Path(path).Methods(methods...).Handler(handler)
		}
	}
}
