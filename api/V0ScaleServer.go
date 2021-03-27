package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	mesosproto "mesos-k3s/proto"

	mesos "mesos-k3s/mesos"
)

// V0ScaleK3S will scale the k3s service
// example:
// curl -X GET 127.0.0.1:10000/v0/k3s/scale/{count of instances} -d 'JSON'
func V0ScaleK3S(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	d := []byte("nok")

	if vars["count"] != "" {
		newCount, _ := strconv.Atoi(vars["count"])
		oldCount := config.K3SServerMax
		logrus.Debug("V0ScaleK3S: oldCount: ", oldCount)
		config.K3SServerMax = newCount
		i := (newCount - oldCount)
		// change the number to be positiv
		if i < 0 {
			i = i * -1
		}

		// Scale Up
		if newCount > oldCount {
			logrus.Info("K3S Scale Up ", i)
			revive := &mesosproto.Call{
				Type: mesosproto.Call_REVIVE,
			}
			mesos.Call(revive)
		}

		// Scale Down
		if newCount < oldCount {
			logrus.Info("K3S Scale Down ", i)

			for x := newCount; x < oldCount; x++ {
				task := mesos.StatusK3SServer(x)
				id := task.Status.TaskID.Value
				ret := mesos.Kill(id)

				logrus.Info("V0TaskKill: ", ret)
				config.K3SServerCount--
			}
		}

		d = []byte(strconv.Itoa(newCount - oldCount))
	}

	logrus.Debug("HTTP GET V0ScaleK3S: ", string(d))

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
