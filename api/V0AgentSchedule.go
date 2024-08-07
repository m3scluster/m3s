package api

import (
	"net/http"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// V0AgentSchedule - Set Agents to schedule
// example:
// curl -X PUT 127.0.0.1:10000/api/m3s/v0/agent/schedule
func (e *API) V0AgentSchedule(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0AgentSchedule").Debug("Set Schedule")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	e.Kubernetes.SetSchedule()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write([]byte("Set Agent Schedule"))
}
