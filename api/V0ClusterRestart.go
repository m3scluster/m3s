package api

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// V0ClusterRestart - Restart the cluster
// example:
// curl -X PUT 127.0.0.1:10000/v0/cluster/restart
func (e *API) V0ClusterRestart(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0ClusterRestart").Debug("Restart Cluster")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if e.Config.DSMax == 0 && e.Config.K3SServerMax == 0 && e.Config.K3SAgentMax == 0 {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	e.ClusterRestart()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
}

// ClusterRestart will scale down all K8 instances and scale up again.
func (e *API) ClusterRestart() {
	e.clusterStop(false)

	ticker := time.NewTicker(e.Config.EventLoopTime)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		if e.Redis.CountRedisKey(e.Framework.FrameworkName+":server:*", "") == 0 &&
			e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", "") == 0 {
			logrus.WithField("func", "api.V0ClusterRestart").Debug("All services down")
			goto start
		}
	}

start:
	ticker.Stop()
	e.clusterStart()
}
