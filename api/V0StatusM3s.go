package api

import (
	"context"
	"encoding/json"
	"net/http"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	"github.com/AVENTER-UG/mesos-m3s/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// V0StatusM3s gives out the current status of the M3s services
// example:
// curl -X GET 127.0.0.1:10000/api/m3s/v0/status/m3s
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

	nodes, err := e.Kubernetes.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.WithField("func", "api.getStatus").Error("Could not get nodes from controller: ", err.Error())
	}

	for _, node := range nodes.Items {
		taskID := e.Kubernetes.GetTaskIDFromLabel(node.Labels)
		if taskID == "" {
			taskID = e.Kubernetes.GetTaskIDFromAnnotation(node.Annotations)
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
