package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0ScaleDatastore will scale the k3s agent service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/etcd/scale/{count of instances} -d 'JSON'
func (e *API) V0ScaleDatastore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := e.CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}
	d := e.ErrorMessage(0, "V0ScaleDatastore", "ok")

	if vars["count"] != "" {
		newCount, _ := strconv.Atoi(vars["count"])
		oldCount := e.Config.DSMax
		logrus.Debug("V0ScaleDatastore: oldCount: ", oldCount)
		e.Config.DSMax = newCount

		d = []byte(strconv.Itoa(newCount - oldCount))

		// Save current config
		e.Redis.SaveConfig(*e.Config)

		// if scale down, kill not needes agents
		keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":datastore:*")

		for keys.Next(e.Redis.CTX) {
			key := e.Redis.GetRedisKey(keys.Val())
			task := e.Mesos.DecodeTask(key)
			task.Instances = newCount
			e.Redis.SaveTaskRedis(task)

			if newCount < oldCount {
				e.Mesos.Kill(task.TaskID, task.Agent)
				logrus.Debug("V0ScaleDatastore: ", task.TaskID)
			}
			if newCount > oldCount {
				e.Mesos.Revive()
			}
			oldCount = oldCount - 1
		}
	}

	logrus.Debug("HTTP GET V0ScaleDatastore: ", string(d))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}
