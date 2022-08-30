package mesos

import (
	"time"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	"github.com/sirupsen/logrus"
)

var reviveLock bool

// Heartbeat function for mesos
func (e *Scheduler) Heartbeat() {
	// Check Connection state of Redis
	err := e.API.PingRedis()
	if err != nil {
		e.API.ConnectRedis()
	}

	dsState := e.healthCheckDatastore()
	k3sState := e.healthCheckK3s()
	k3sAgenteState := e.healthCheckAgent()

	// if DataStorage container is not running or unhealthy, fix it.
	if !dsState {
		go e.scheduleRevive()
		e.StartDatastore("")
	}

	// if Datastorage is running and K3s not, deploy K3s
	if dsState && !k3sState {
		go e.scheduleRevive()
		e.StartK3SServer("")
	}

	// if k3s is running, deploy the agent
	if k3sState && !k3sAgenteState {
		go e.scheduleRevive()
		e.StartK3SAgent("")
	}

	if k3sState && k3sAgenteState && dsState {
		mesosutil.SuppressFramework()
	}
}

// CheckState check the current state of every task
func (e *Scheduler) CheckState() {
	keys := e.API.GetAllRedisKeys(e.Framework.FrameworkName + ":*")

	for keys.Next(e.API.Redis.RedisCTX) {
		// get the values of the current key
		key := e.API.GetRedisKey(keys.Val())

		if e.API.CheckIfNotTask(keys) {
			continue
		}

		task := mesosutil.DecodeTask(key)

		if task.TaskID == "" || task.TaskName == "" {
			continue
		}

		if task.State == "TASK_FINISHED" {
			e.API.DelRedisKey(key)
		}

		if task.State == "" && e.API.CountRedisKey(task.TaskName+":*") <= task.Instances {
			go e.scheduleRevive()
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
	}
}

// scheduleRevive - Schedule Revive Tasks
func (e *Scheduler) scheduleRevive() {
	if reviveLock {
		return
	}
	logrus.WithField("func", "mesos.scheduleRevive").Debug("Schedule Revive")
	reviveLock = true

	select {
	case <-time.After(e.Config.ReviveLoopTime):
		mesosutil.Revive()
	}
	reviveLock = false
}

// HeartbeatLoop - The main loop for the hearbeat
func (e *Scheduler) HeartbeatLoop() {
	ticker := time.NewTicker(e.Config.EventLoopTime)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		e.Heartbeat()
		e.CheckState()
	}
}
