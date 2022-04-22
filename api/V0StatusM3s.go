package api

import (
	"encoding/json"
	"net/http"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0StatusM3s gives out the current status of the M3s services
// example:
// curl -X GET 127.0.0.1:10000/v0/status/m3s -d 'JSON'
func V0StatusM3s(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	getStatus()
	status := config.M3SStatus

	d, _ := json.Marshal(&status)

	logrus.Debug("HTTP GET V0StatusM3s: ", string(d))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}

func getStatus() {
	config.M3SStatus.Etcd = map[string]string{}
	config.M3SStatus.Agent = map[string]string{}
	config.M3SStatus.Server = map[string]string{}
	services := []string{"etcd", "server", "agent"}

	for _, service := range services {
		keys := GetAllRedisKeys(framework.FrameworkName + ":" + service + ":*")

		for keys.Next(config.RedisCTX) {
			key := GetRedisKey(keys.Val())
			var task mesosutil.Command
			json.Unmarshal([]byte(key), &task)

			if service == "etcd" {
				config.M3SStatus.Etcd[task.TaskID] = task.State
			} else if service == "agent" {
				config.M3SStatus.Agent[task.TaskID] = task.State
			} else if service == "server" {
				config.M3SStatus.Server[task.TaskID] = task.State
			}
		}
	}
}
