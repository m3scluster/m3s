package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0ScaleK3SAgent will scale the k3s agent service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/agent/scale/{count of instances} -d 'JSON'
func (e *API) V0ScaleK3SAgent(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0ScaleK3SAgent").Debug("Call")

	vars := mux.Vars(r)

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if vars == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	d := e.ErrorMessage(0, "V0ScaleK3SAgent", "ok")

	if vars["count"] != "" {
		count, err := strconv.Atoi(vars["count"])
		if err != nil {
			logrus.WithField("func", "api.V0ScaleK3SAgent").Error("Error: ", err.Error())
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		d = e.scaleAgent(count)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}

// scaleAgent - can scale up and down the K8 worker nodes
func (e *API) scaleAgent(count int) []byte {
	r := e.scale(count, e.Config.K3SAgentMax, ":agent:*")
	e.Config.K3SAgentMax = count
	return r
}
