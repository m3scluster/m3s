package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	mesosutil "github.com/AVENTER-UG/mesos-util"
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
		e.SaveConfig()

		// if scale down, kill not needes agents
		if newCount < oldCount {
			keys := e.GetAllRedisKeys(e.Framework.FrameworkName + ":datastore:*")

			for keys.Next(e.Redis.RedisCTX) {
				if newCount < oldCount {
					key := e.GetRedisKey(keys.Val())

					var task mesosutil.Command
					json.Unmarshal([]byte(key), &task)
					mesosutil.Kill(task.TaskID, task.Agent)
					logrus.Debug("V0ScaleDatastore: ", task.TaskID)
				}
				oldCount = oldCount - 1
			}
		}
	}

	logrus.Debug("HTTP GET V0ScaleDatastore: ", string(d))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
