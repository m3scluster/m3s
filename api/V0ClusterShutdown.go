package api

import (
	"net/http"

	"github.com/sirupsen/logrus"
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

	// Save current amount of services for the case of restart but only
	// if the amount is not 0
	if e.Config.DSMax != 0 && e.Config.K3SServerMax != 0 && e.Config.K3SAgentMax != 0 {
		e.DSMaxRestore = e.Config.DSMax
		e.K3SServerMaxRestore = e.Config.K3SServerMax
		e.K3SAgentMaxRestore = e.Config.K3SAgentMax
	}

	e.scaleAgent(0)
	e.scaleServer(0)
	e.scaleDatastore(0)

	e.Mesos.SuppressFramework()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
}
