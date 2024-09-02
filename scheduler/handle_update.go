package scheduler

import (
	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// HandleUpdate will handle the offers event of mesos
func (e *Scheduler) HandleUpdate(event *mesosproto.Event) error {
	update := event.Update

	msg := &mesosproto.Call{
		Type: mesosproto.Call_ACKNOWLEDGE.Enum(),
		Acknowledge: &mesosproto.Call_Acknowledge{
			AgentId: update.Status.GetAgentId(),
			TaskId:  update.Status.GetTaskId(),
			Uuid:    update.Status.GetUuid(),
		},
	}

	// get the task of the current event, change the state and get some info's we need for later use
	task := e.Redis.GetTaskFromEvent(update)

	// if these object have not TaskID it's currently unknown by these framework.
	if task.TaskID == "" {
		logrus.WithField("func", "scheduler.HandleUpdate").Debug("Could not found Task in Redis: ", update.Status.GetTaskId())

		if *update.Status.State != mesosproto.TaskState_TASK_LOST {
			e.Mesos.Kill(*update.Status.GetTaskId().Value, *update.Status.GetAgentId().Value)
		}
	}

	task.State = update.Status.State.String()
	switch *update.Status.State {
	case mesosproto.TaskState_TASK_FAILED, mesosproto.TaskState_TASK_KILLED, mesosproto.TaskState_TASK_LOST, mesosproto.TaskState_TASK_ERROR, mesosproto.TaskState_TASK_FINISHED:
		logrus.WithField("func", "scheduler.HandleUpdate").Warn("Task State: " + task.State + " " + task.TaskID + " (" + task.TaskName + ")")
		e.Redis.DelRedisKey(task.TaskName + ":" + task.TaskID)
		// remove unready K8 node from redis
		if task.TaskName == e.Framework.FrameworkName+":server" {
			node := e.Kubernetes.GetK8NodeFromTask(*task)
			e.Redis.DelRedisKey(e.Framework.FrameworkName + ":kubernetes:" + node.Name)
		}
		if task.TaskName == e.Framework.FrameworkName+":agent" {
			node := e.Kubernetes.GetK8NodeFromTask(*task)
			e.Kubernetes.DeleteNode(node.Name)
			e.Kubernetes.SetUnschedule()
			e.Redis.DelRedisKey(e.Framework.FrameworkName + ":kubernetes:" + node.Name)
		}

		return e.Mesos.Call(msg)
	case mesosproto.TaskState_TASK_RUNNING:
		logrus.WithField("func", "scheduler.HandleUpdate").Info("Task State: " + task.State + " " + task.TaskID + " (" + task.TaskName + ")")
		task.MesosAgent = e.Mesos.GetAgentInfo(*update.Status.GetAgentId().Value)
		task.NetworkInfo = e.Mesos.GetNetworkInfo(task.TaskID)
		task.Agent = update.Status.AgentId.GetValue()
		// remember information for the boostrap server to reach it later
		if task.TaskName == e.Framework.FrameworkName+":server" {
			// if the framework is running as container, and the task hostname is the same like the frameworks one,
			if e.Config.DockerRunning && (task.MesosAgent.Hostname == e.Config.Hostname) {
				e.Config.K3SServerHostname = task.Hostname
			}
			e.Config.K3SServerPort = int(task.DockerPortMappings[0].GetHostPort())
		}
	}

	e.Redis.SaveTaskRedis(task)
	return e.Mesos.Call(msg)
}
