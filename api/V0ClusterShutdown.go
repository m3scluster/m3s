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
		return
	}

	e.scaleAgent(0)
	e.scaleServer(0)
	e.scaleDatastore(0)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
}
