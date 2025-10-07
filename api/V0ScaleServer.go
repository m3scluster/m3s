package api

import (
	"net/http"
	"strconv"
	"time"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	"github.com/gorilla/mux"
)

// V0ScaleK3SServer will scale the k3s server service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/server/scale/{count of instances} -d 'JSON'
func (e *API) V0ScaleK3SServer(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0ScaleK3SServer").Debug("Call")

	vars := mux.Vars(r)

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if vars == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	d := e.ErrorMessage(0, "V0ScaleK3SServer", "ok")

	if vars["count"] != "" {
		count, err := strconv.Atoi(vars["count"])
		if err != nil {
			logrus.WithField("func", "api.V0K3SServer").Error("Error: ", err.Error())
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		d = e.scaleServer(count)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}

func (e *API) scaleServer(count int) []byte {
	r := e.Scale(count, e.Config.K3SServerMax, "server")
	e.Config.K3SServerMax = count
	return r
}

func (e *API) Scale(newCount int, oldCount int, key string) []byte {
	logrus.WithField("func", "api.scale").Debug("Scale " + key + " current: " + strconv.Itoa(oldCount) + " new: " + strconv.Itoa(newCount))

	d := []byte(strconv.Itoa(newCount - oldCount))

	// Save current config
	e.Redis.SaveConfig(*e.Config)

	// if scale down, kill not needes agents
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":" + key + ":*")

	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		task := e.Mesos.DecodeTask(key)
		task.Instances = newCount
		e.Redis.SaveTaskRedis(task)

		if newCount < oldCount {
			node := e.Kubernetes.GetK8NodeFromTask(*task)
			if task.TaskName == e.Framework.FrameworkName+":agent" && node.Name != "" {
				// remove the agent nodes before we kill the agent
				e.Kubernetes.DeleteNode(node.Name)
				logrus.WithField("func", "api.scale").Debug("Delete Node: ", task.TaskID)
				time.Sleep(5 * time.Second)
			}
			e.Mesos.Kill(task.TaskID, task.Agent)
			logrus.WithField("func", "api.scale").Debug("Scale Down TaskID: ", task.TaskID)
		}
		if newCount > oldCount {
			e.Mesos.Revive()
		}
		oldCount = oldCount - 1
	}

	return d
}

// Server will scale down all K8 servers & agents instances and scale up again.
func (e *API) ServerRestart() {
	e.serverAndAgentStop()

	ticker := time.NewTicker(e.Config.EventLoopTime)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		if e.Redis.CountRedisKey(e.Framework.FrameworkName+":server:*", "") == 0 &&
			e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", "") == 0 {
			logrus.WithField("func", "api.V0ServerRestart").Debug("All services down")
			goto start
		}
	}

start:
	ticker.Stop()
	e.serverAndAgentStart()
}
