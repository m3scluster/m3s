package api

import (
	"net/http"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// V0StatusK8s gives out the current status of the K8s services
// example:
// curl -X GET 127.0.0.1:10000/v0/status/k8s
func (e *API) V0StatusK8s(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0StatusK3s").Debug("Call")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	if e.K3SAgentStatus {
		w.Write([]byte("ok"))
	} else {
		w.Write([]byte("nok"))
	}
}
