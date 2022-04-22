package mesos

import (
	"encoding/json"
	"time"

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

		if task.TaskID == "" || task.TaskName == "" {
			continue
		}

		if task.State == "" {
			mesosutil.Revive()
			task.State = "__NEW"
			// these will save the current time at the task. we need it to check
			// if the state will change in the next 'n min. if not, we have to
			// give these task a recall.
			task.StateTime = time.Now()

			// add task to communication channel
			framework.CommandChan <- task

			data, _ := json.Marshal(task)
			err := config.RedisClient.Set(config.RedisCTX, task.TaskName+":"+task.TaskID, data, 0).Err()
			if err != nil {
				logrus.Error("HandleUpdate Redis set Error: ", err)
			}

			logrus.Info("Scheduled Mesos Task: ", task.TaskName)
		}

		if task.State == "__NEW" {
			suppress = false
			config.Suppress = false
		}
	}

	if suppress && !config.Suppress {
		mesosutil.SuppressFramework()
		config.Suppress = true
	}
}

// K3SHeartbeat to execute K3S Bootstrap API Server commands
func K3SHeartbeat() {
	if api.CountRedisKey(framework.FrameworkName+":etcd:*") < config.ETCDMax {
		StartEtcd("")
	}
	if getEtcdStatus() == "TASK_RUNNING" && !IsK3SServerRunning() {
		if api.CountRedisKey(framework.FrameworkName+":server:*") < config.K3SServerMax {
			StartK3SServer("")
		}
	}
	if IsK3SServerRunning() {
		if api.CountRedisKey(framework.FrameworkName+":agent:*") < config.K3SAgentMax {
			StartK3SAgent("")
		}
	}
}
