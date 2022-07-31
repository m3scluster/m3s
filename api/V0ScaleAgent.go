package api

import (
	"net/http"
	"strconv"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0ScaleK3SAgent will scale the k3s agent service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/agent/scale/{count of instances} -d 'JSON'
func (e *API) V0ScaleK3SAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := e.CheckAuth(r, w)

	if vars == nil || !auth {
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
		e.SaveConfig()

		// if scale down, kill not needes agents
		if newCount < oldCount {
			keys := e.GetAllRedisKeys(e.Framework.FrameworkName + ":agent:*")

			for keys.Next(e.Redis.RedisCTX) {
				if newCount < oldCount {
					key := e.GetRedisKey(keys.Val())

					task := mesosutil.DecodeTask(key)

					mesosutil.Kill(task.TaskID, task.MesosAgent.Slaves[0].ID)
					logrus.Debug("V0ScaleK3SAgent: ", task.TaskID)
				}
				oldCount = oldCount - 1
			}
		}
	}

	logrus.Debug("HTTP GET V0ScaleK3SAgent: ", string(d))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
