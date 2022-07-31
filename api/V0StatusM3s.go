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
func (e *API) V0StatusM3s(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := e.CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	e.getStatus()
	status := e.Config.M3SStatus

	d, _ := json.Marshal(&status)

	logrus.Debug("HTTP GET V0StatusM3s: ", string(d))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}

func (e *API) getStatus() {
	e.Config.M3SStatus.Etcd = map[string]string{}
	e.Config.M3SStatus.Agent = map[string]string{}
	e.Config.M3SStatus.Server = map[string]string{}
	services := []string{"etcd", "server", "agent"}

	for _, service := range services {
		keys := e.GetAllRedisKeys(e.Framework.FrameworkName + ":" + service + ":*")

		for keys.Next(e.Redis.RedisCTX) {
			key := e.GetRedisKey(keys.Val())
			task := mesosutil.DecodeTask(key)

			if service == "etcd" {
				e.Config.M3SStatus.Etcd[task.TaskID] = task.State
			} else if service == "agent" {
				e.Config.M3SStatus.Agent[task.TaskID] = task.State
			} else if service == "server" {
				e.Config.M3SStatus.Server[task.TaskID] = task.State
			}
		}
	}
}
