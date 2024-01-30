package api

import (
	"net/http"
	"strings"
	"time"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// V0Restart - Restart cluster or datastore, server or agent.
// example:
// curl -X PUT 127.0.0.1:10000/v0/cluster/restart
func (e *API) V0Restart(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0Restart").Debug("Restart Cluster")

	pathFragments := strings.Split(r.URL.Path, "/")

	component := pathFragments[4]

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if e.Config.DSMax == 0 && e.Config.K3SServerMax == 0 && e.Config.K3SAgentMax == 0 {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	switch component {
	case "cluster":
		e.ClusterRestart(false)
	case "ds":
		e.ClusterRestart(true)
	case "server":
		e.ServerRestart()
	case "agent":
		e.AgentRestart()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write([]byte("Restart Scheduled."))
}

// ClusterRestart will scale down all K8 instances and scale up again.
func (e *API) ClusterRestart(ds bool) {
	e.clusterStop(ds)

	logrus.WithField("func", "api.V0Restart").Debug("All services scheduled for stop.")
	ticker := time.NewTicker(e.Config.EventLoopTime)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		if e.Redis.CountRedisKey(e.Framework.FrameworkName+":server:*", "") == 0 &&
			e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", "") == 0 &&
			(!ds || e.Redis.CountRedisKey(e.Framework.FrameworkName+":datastore:*", "") == 0) {
			logrus.WithField("func", "api.V0Restart").Debug("All services down")
			goto start
		}
	}

start:
	ticker.Stop()
	e.clusterStart()
}
