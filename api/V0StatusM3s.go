package api

import (
	"encoding/json"
	"net/http"
	"strings"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	"github.com/AVENTER-UG/mesos-m3s/types"
	corev1 "k8s.io/api/core/v1"
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
	e.Config.M3SStatus.Agent = map[string]types.K3SNodeStatus{}
	e.Config.M3SStatus.Server = map[string]types.K3SNodeStatus{}
	services := []string{"datastore", "server", "agent"}

	K3sNodeTaskMap := map[string]string{}

	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":kubernetes:*")
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		var node corev1.Node
		err := json.NewDecoder(strings.NewReader(key)).Decode(&node)
		if err != nil {
			logrus.WithField("func", "scheduler.V0GetKubeVersion").Error("Could not decode kubernetes node: ", err.Error())
			continue
		}

		taskID := e.GetTaskIDFromLabel(node.Labels)
		if taskID == "" {
			taskID = e.GetTaskIDFromAnnotation(node.Annotations)
		}

		K3sNodeTaskMap[taskID] = node.Name
	}

	for _, service := range services {
		keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":" + service + ":*")

		for keys.Next(e.Redis.CTX) {
			key := e.Redis.GetRedisKey(keys.Val())
			task := e.Mesos.DecodeTask(key)

			if service == "datastore" {
				e.Config.M3SStatus.Datastore[task.TaskID] = task.State
			} else if service == "agent" {
				agent := types.K3SNodeStatus{}
				agent.Status = task.State
				agent.K3sNodeName = K3sNodeTaskMap[task.TaskID]
				e.Config.M3SStatus.Agent[task.TaskID] = agent

			} else if service == "server" {
				server := types.K3SNodeStatus{}
				server.Status = task.State
				server.K3sNodeName = K3sNodeTaskMap[task.TaskID]
				e.Config.M3SStatus.Server[task.TaskID] = server
			}
		}
	}
}

// GetTaskIDFromAnnotation will return the Mesos Task ID in the annotation string
func (e *API) GetTaskIDFromAnnotation(annotations map[string]string) string {
	for i, annotation := range annotations {
		if i == "k3s.io/node-args" {
			var args []string
			err := json.Unmarshal([]byte(annotation), &args)
			if err != nil {
				logrus.WithField("func", "scheduler.GetTaskIDFromAnnotation").Error("Could not decode kubernetes node annotation: ", err.Error())
				continue
			}
			for _, arg := range args {
				if strings.Contains(arg, "taskid") {
					value := strings.Split(arg, "=")
					if len(value) == 2 {
						return value[1]
					}
				}
			}
		}
	}
	return ""
}

// GetTaskIDFromLabel will return the Mesos Task ID in the label string
func (e *API) GetTaskIDFromLabel(labels map[string]string) string {
	for i, label := range labels {
		if i == "m3s.aventer.biz/taskid" {
			return label
		}
	}
	return ""
}
