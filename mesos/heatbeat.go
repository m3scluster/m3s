package mesos

import (
	"encoding/json"

	api "github.com/AVENTER-UG/mesos-m3s/api"
	mesosutil "github.com/AVENTER-UG/mesos-util"
	"github.com/sirupsen/logrus"
)

// Heartbeat function for mesos
func Heartbeat() {
	K3SHeartbeat()
	keys := api.GetAllRedisKeys("*")
	suppress := true

	for keys.Next(config.RedisCTX) {
		// get the values of the current key
		key := api.GetRedisKey(keys.Val())

		var task mesosutil.Command
		json.Unmarshal([]byte(key), &task)

		if task.TaskID == "" {
			continue
		}

		if task.State == "" {
			framework.CommandChan <- task

			task.State = "__NEW"
			data, _ := json.Marshal(task)
			err := config.RedisClient.Set(config.RedisCTX, task.TaskName+":"+task.TaskID, data, 0).Err()
			if err != nil {
				logrus.Error("HandleUpdate Redis set Error: ", err)
			}
			logrus.Info("Scheduled Mesos Task: ", task.TaskName)
		}

		if task.State == "__NEW" {
			mesosutil.Revive()
			suppress = false
		}
	}

	if suppress {
		mesosutil.SuppressFramework()
	}
}

// K3SHeartbeat to execute K3S Bootstrap API Server commands
func K3SHeartbeat() {
	if api.CountRedisKey(config.PrefixHostname+"etcd:*") < config.ETCDMax {
		StartEtcd("")
	}
	if getEtcdStatus() == "TASK_RUNNING" && !IsK3SServerRunning() {
		if api.CountRedisKey(config.PrefixHostname+"server:*") < config.K3SServerMax {
			StartK3SServer("")
		}
	}
	if IsK3SServerRunning() {
		if api.CountRedisKey(config.PrefixHostname+"agent:*") < config.K3SAgentMax {
			StartK3SAgent("")
		}
	}
}
