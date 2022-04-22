package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0ScaleEtcd will scale the k3s agent service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/etcd/scale/{count of instances} -d 'JSON'
func V0ScaleEtcd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}
	d := ErrorMessage(0, "V0ScaleEtcd", "ok")

	if vars["count"] != "" {
		newCount, _ := strconv.Atoi(vars["count"])
		oldCount := config.ETCDMax
		logrus.Debug("V0ScaleEtcd: oldCount: ", oldCount)
		config.ETCDMax = newCount

		d = []byte(strconv.Itoa(newCount - oldCount))

		// Save current config
		SaveConfig()

		// if scale down, kill not needes agents
		if newCount < oldCount {
			keys := GetAllRedisKeys(framework.FrameworkName + ":etcd:*")

			for keys.Next(config.RedisCTX) {
				if newCount < oldCount {
					key := GetRedisKey(keys.Val())

					var task mesosutil.Command
					json.Unmarshal([]byte(key), &task)
					mesosutil.Kill(task.TaskID, task.Agent)
					logrus.Debug("V0ScaleEtcd: ", task.TaskID)
				}
				oldCount = oldCount - 1
			}
		}
	}

	logrus.Debug("HTTP GET V0ScaleEtcd: ", string(d))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
