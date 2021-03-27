package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	mesos "mesos-k3s/mesos"
)

// V0KillTask will kill the given task id
// example:
// curl -X GET 127.0.0.1:10000/v0/task/kill/{count of instances} -d 'JSON'
func V0KillTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	d := []byte("nok")

	if vars["id"] != "" {
		id := vars["id"]
		ret := mesos.Kill(id)

		logrus.Error("V0TaskKill: ", ret)

		d = []byte("ok")
	}

	logrus.Debug("HTTP GET V0TaskKill: ", string(d))

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
