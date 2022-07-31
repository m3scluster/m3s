package mesos

import (
	"time"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	"github.com/sirupsen/logrus"
)

// Heartbeat function for mesos
func (e *Scheduler) Heartbeat() {
	// Check Connection state of Redis
	err := e.API.PingRedis()
	if err != nil {
		e.API.ConnectRedis()
	}

	etcdState := e.healthCheckETCD()
	k3sState := e.healthCheckK3s()

	// if ETCD is not running or unhealthy, fix it.
	if !etcdState && e.API.CountRedisKey(e.Framework.FrameworkName+":etcd:*") < e.Config.ETCDMax {
		e.StartEtcd("")
	}

	// if ETCD is running and K3s not, deploy K3s
	if etcdState && !k3sState {
		if e.API.CountRedisKey(e.Framework.FrameworkName+":server:*") < e.Config.K3SServerMax {
			e.StartK3SServer("")
		}
	}

	// if k3s is running, deploy the agent
	if k3sState {
		if e.API.CountRedisKey(e.Framework.FrameworkName+":agent:*") < e.Config.K3SAgentMax {
			e.StartK3SAgent("")
		}
	}
}

// CheckState check the current state of every task
func (e *Scheduler) CheckState() {
	keys := e.API.GetAllRedisKeys(e.Framework.FrameworkName + ":*")
	suppress := true

	for keys.Next(e.API.Redis.RedisCTX) {
		// get the values of the current key
		key := e.API.GetRedisKey(keys.Val())

		task := mesosutil.DecodeTask(key)

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
			e.Framework.CommandChan <- task

			e.API.SaveTaskRedis(task)

			logrus.Info("Scheduled Mesos Task: ", task.TaskName)
		}

		if task.State == "__NEW" {
			suppress = false
			e.Config.Suppress = false
		}
	}

	if suppress && !e.Config.Suppress {
		mesosutil.SuppressFramework()
		e.Config.Suppress = true
	}
}

// HeartbeatLoop - The main loop for the hearbeat
func (e *Scheduler) HeartbeatLoop() {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		e.Heartbeat()
		e.CheckState()
	}
}
