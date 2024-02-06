package api

import (
	"net/http"
	"strconv"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	"github.com/gorilla/mux"
)

// V0ScaleDatastore will scale the k3s agent service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/etcd/scale/{count of instances} -d 'JSON'
func (e *API) V0ScaleDatastore(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0ScaleDatastore").Debug("Call")

	vars := mux.Vars(r)

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if vars == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	d := e.ErrorMessage(0, "V0ScaleDatastore", "ok")

	if vars["count"] != "" {
		count, err := strconv.Atoi(vars["count"])
		if err != nil {
			logrus.WithField("func", "api.V0ScaleDatastore").Error("Error: ", err.Error())
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		d = e.scaleDatastore(count)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}

// scaleDatastore - can scale up and down the K8 datastore
func (e *API) scaleDatastore(count int) []byte {
	r := e.scale(count, e.Config.DSMax, "datastore")
	e.Config.DSMax = count
	return r
}
