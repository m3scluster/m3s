package api

import (
	"net/http"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// V0UnCordonK3SAgent will uncordon the K3S nodes if they are manually cordoned.
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/agent/uncordon'
func (e *API) V0UnCordonK3SAgent(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0UnCordonK3SAgent").Debug("Uncordon K3S Agents")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Save current amount of services for the case of restart but only
	// if the amount is not 0
	e.Config.K3SDisableScheduling = false
	e.Redis.SaveConfig(*e.Config)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write([]byte("Nodes Uncordoned, Scheduling will now resume"))
}

// V0CordonK3SAgent will cordon the K3S nodes manually.
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/agent/cordon'
func (e *API) V0CordonK3SAgent(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0CordonK3SAgent").Debug("Cordon K3S Agents")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Save current amount of services for the case of restart but only
	// if the amount is not 0
	e.Config.K3SDisableScheduling = true
	e.Redis.SaveConfig(*e.Config)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write([]byte("Nodes Cordoned, Scheduling will be disabled"))
}
