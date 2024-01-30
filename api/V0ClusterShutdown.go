package api

import (
	"net/http"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// V0ClusterShutdown - Shutdown the whole K8 cluster
// example:
// curl -X PUT 127.0.0.1:10000/v0/cluster/shutdown'
func (e *API) V0ClusterShutdown(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0ClusterShutdown").Debug("Shutdown Cluster")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	e.clusterStop(true)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
}

// clusterStop will scale down the cluster
//
//	bool ds = scale down the datastore too
func (e *API) clusterStop(ds bool) {
	logrus.WithField("func", "api.clusterStop").Debug("Shutdown Cluster")

	// Save current amount of services for the case of restart but only
	// if the amount is not 0
	if e.Config.DSMax != 0 && e.Config.K3SServerMax != 0 && e.Config.K3SAgentMax != 0 {
		e.Config.DSMaxRestore = e.Config.DSMax
		e.Config.K3SServerMaxRestore = e.Config.K3SServerMax
		e.Config.K3SAgentMaxRestore = e.Config.K3SAgentMax
	}

	e.Redis.SaveConfig(*e.Config)

	e.scaleAgent(0)
	e.scaleServer(0)
	if ds {
		e.scaleDatastore(0)
	}

	e.Mesos.SuppressFramework()
}

// serverAndAgentStop will scale down only the server & agents
func (e *API) serverAndAgentStop() {
	// Save current amount of services for the case of restart but only
	// if the amount is not 0
	if e.Config.K3SServerMax != 0 && e.Config.K3SAgentMax != 0 {
		e.Config.K3SServerMaxRestore = e.Config.K3SServerMax
		e.Config.K3SAgentMaxRestore = e.Config.K3SAgentMax
	}

	e.Redis.SaveConfig(*e.Config)

	e.scaleAgent(0)
	e.scaleServer(0)

}
