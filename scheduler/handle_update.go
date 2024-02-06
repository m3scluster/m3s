package scheduler

import (
	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// HandleUpdate will handle the offers event of mesos
func (e *Scheduler) HandleUpdate(event *mesosproto.Event) error {
	update := event.Update

	if update.Status.UUID == nil {
		return nil
	}

	msg := &mesosproto.Call{
		Type: mesosproto.Call_ACKNOWLEDGE,
		Acknowledge: &mesosproto.Call_Acknowledge{
			AgentID: *update.Status.AgentID,
			TaskID:  update.Status.TaskID,
			UUID:    update.Status.UUID,
		},
	}

	// get the task of the current event, change the state and get some info's we need for later use
	task := e.Redis.GetTaskFromEvent(update)

	// if these object have not TaskID it's currently unknown by these framework.
	if task.TaskID == "" {
		logrus.WithField("func", "scheduler.HandleUpdate").Debug("Could not found Task in Redis: ", update.Status.TaskID.Value)

		if *update.Status.State != mesosproto.TASK_LOST {
			e.Mesos.Kill(update.Status.TaskID.Value, update.Status.AgentID.Value)
		}
	}

	task.State = update.Status.State.String()

	switch *update.Status.State {
	case mesosproto.TASK_FAILED, mesosproto.TASK_KILLED, mesosproto.TASK_LOST, mesosproto.TASK_ERROR, mesosproto.TASK_FINISHED:
		logrus.WithField("func", "scheduler.HandleUpdate").Warn("Task State: " + task.State + " " + task.TaskID + " (" + task.TaskName + ")")
		e.Redis.DelRedisKey(task.TaskName + ":" + task.TaskID)
		// remove unready K8 node from redis
		if task.TaskName == e.Framework.FrameworkName+":agent" {
			node := e.getK8NodeFromTask(task)
			e.Redis.DelRedisKey(e.Framework.FrameworkName + ":kubernetes:" + node.Name)
		}

		return e.Mesos.Call(msg)
	case mesosproto.TASK_RUNNING:
		logrus.WithField("func", "scheduler.HandleUpdate").Info("Task State: " + task.State + " " + task.TaskID + " (" + task.TaskName + ")")
		task.MesosAgent = e.Mesos.GetAgentInfo(update.Status.GetAgentID().Value)
		task.NetworkInfo = e.Mesos.GetNetworkInfo(task.TaskID)
		task.Agent = update.Status.GetAgentID().Value
		// remember information for the boostrap server to reach it later
		if task.TaskName == e.Framework.FrameworkName+":server" {
			// if the framework is running as container, and the task hostname is the same like the frameworks one,
			if e.Config.DockerRunning && (task.MesosAgent.Hostname == e.Config.Hostname) {
				e.Config.K3SServerHostname = task.Hostname
			}
			e.Config.K3SServerPort = int(task.DockerPortMappings[0].HostPort)
		}
	}

	e.Redis.SaveTaskRedis(task)
	return e.Mesos.Call(msg)
}
