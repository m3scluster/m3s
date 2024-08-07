package api

import (
	"net/http"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// V0AgentUnSchedule - Set Agent to UnSchedule
// example:
// curl -X DELETE 127.0.0.1:10000/api/m3s/v0/agent/schedule
func (e *API) V0AgentUnschedule(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0AgentUnschedule").Debug("Set Unschedule")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	e.Kubernetes.SetUnschedule()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write([]byte("Set Agent Unschedule"))
}
