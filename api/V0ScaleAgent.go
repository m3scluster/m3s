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
	vars := mux.Vars(r)

	if vars == nil || !e.CheckAuth(r, w) {
		return
	}

	d := e.ErrorMessage(0, "V0ScaleK3SAgent", "ok")

	if vars["count"] != "" {
		newCount, _ := strconv.Atoi(vars["count"])
		oldCount := e.Config.K3SAgentMax
		logrus.Debug("V0ScaleK3SAgent: oldCount: ", oldCount)
		e.Config.K3SAgentMax = newCount

		d = []byte(strconv.Itoa(newCount - oldCount))

		// Save current config
		e.Redis.SaveConfig(*e.Config)

		// if scale down, kill not needes agents
		keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":agent:*")

		for keys.Next(e.Redis.CTX) {
			key := e.Redis.GetRedisKey(keys.Val())
			task := e.Mesos.DecodeTask(key)
			task.Instances = newCount
			e.Redis.SaveTaskRedis(task)

			if newCount < oldCount {
				e.Mesos.Kill(task.TaskID, task.MesosAgent.ID)
				logrus.Debug("V0ScaleK3SAgent: ", task.TaskID)
			}
			if newCount > oldCount {
				e.Mesos.Revive()
			}
			oldCount = oldCount - 1
		}
	}

	logrus.Debug("HTTP GET V0ScaleK3SAgent: ", string(d))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}
