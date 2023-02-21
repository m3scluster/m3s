package api

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// V0StatusM3s gives out the current status of the M3s services
// example:
// curl -X GET 127.0.0.1:10000/v0/status/m3s
func (e *API) V0StatusM3s(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0StatusM3s").Debug("Call")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	e.getStatus()
	status := e.Config.M3SStatus

	d, _ := json.Marshal(&status)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}

func (e *API) getStatus() {
	e.Config.M3SStatus.Datastore = map[string]string{}
	e.Config.M3SStatus.Agent = map[string]string{}
	e.Config.M3SStatus.Server = map[string]string{}
	services := []string{"datastore", "server", "agent"}

	for _, service := range services {
		keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":" + service + ":*")

		for keys.Next(e.Redis.CTX) {
			key := e.Redis.GetRedisKey(keys.Val())
			task := e.Mesos.DecodeTask(key)

			if service == "datastore" {
				e.Config.M3SStatus.Datastore[task.TaskID] = task.State
			} else if service == "agent" {
				e.Config.M3SStatus.Agent[task.TaskID] = task.State
			} else if service == "server" {
				e.Config.M3SStatus.Server[task.TaskID] = task.State
			}
		}
	}
}
