package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0ScaleK3SServer will scale the k3s server service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/server/scale/{count of instances} -d 'JSON'
func V0ScaleK3SServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	d := ErrorMessage(0, "V0ScaleK3SServer", "ok")

	if vars["count"] != "" {
		newCount, _ := strconv.Atoi(vars["count"])
		oldCount := config.K3SServerMax
		logrus.Debug("V0ScaleK3SServer: oldCount: ", oldCount)
		config.K3SServerMax = newCount

		d = []byte(strconv.Itoa(newCount - oldCount))

		// Save current config
		SaveConfig()

		// if scale down, kill not needes agents
		if newCount < oldCount {
			keys := GetAllRedisKeys(framework.FrameworkName + ":server:*")

			for keys.Next(config.RedisCTX) {
				if newCount < oldCount {
					key := GetRedisKey(keys.Val())

					var task mesosutil.Command
					json.Unmarshal([]byte(key), &task)
					mesosutil.Kill(task.TaskID, task.Agent)
					logrus.Debug("V0ScaleK3SServer: ", task.TaskID)
				}
				oldCount = oldCount - 1
			}
		}
	}

	logrus.Debug("HTTP GET V0ScaleK3SServer: ", string(d))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
