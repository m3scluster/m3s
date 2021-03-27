package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	mesos "mesos-k3s/mesos"
	mesosproto "mesos-k3s/proto"
)

// V0ScaleK3SAgent will scale the zookeeper service
// example:
// curl -X GET 127.0.0.1:10000/v0/zookeeper/scale/{count of instances} -d 'JSON'
func V0ScaleK3SAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	d := []byte("nok")

	if vars["count"] != "" {
		newCount, _ := strconv.Atoi(vars["count"])
		oldCount := config.K3SAgentMax
		logrus.Debug("V0ScaleK3SAgent: oldCount: ", oldCount)
		config.K3SAgentMax = newCount
		i := (newCount - oldCount)
		// change the number to be positiv
		if i < 0 {
			i = i * -1
		}

		// Scale Up
		if newCount > oldCount {
			logrus.Info("K3SAgent Scale Up ", i)
			revive := &mesosproto.Call{
				Type: mesosproto.Call_REVIVE,
			}
			mesos.Call(revive)
		}

		// Scale Down
		if newCount < oldCount {
			logrus.Info("K3SAgent Scale Down ", i)

			for x := newCount; x < oldCount; x++ {
				task := mesos.StatusK3SAgent(x)
				id := task.Status.TaskID.Value
				ret := mesos.Kill(id)

				logrus.Info("V0TaskKill: ", ret)
				config.K3SAgentCount--
			}
		}

		d = []byte(strconv.Itoa(newCount - oldCount))
	}

	logrus.Debug("HTTP GET V0ScaleK3SAgent: ", string(d))

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
