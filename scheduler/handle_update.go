package scheduler

import (
	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"

	"github.com/sirupsen/logrus"
)

// HandleUpdate will handle the offers event of mesos
func (e *Scheduler) HandleUpdate(event *mesosproto.Event) error {
	logrus.Debug("HandleUpdate")

	update := event.Update

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
		logrus.WithField("func", "scheduler.HandleUpdate").Info("Could not found Task in Redis: ", update.Status.TaskID.Value)
		e.Mesos.Kill(update.Status.TaskID.Value, update.Status.AgentID.Value)
	}

	task.State = update.Status.State.String()

	logrus.WithField("func", "scheduler.HandleUpdate").Debugf("Task State: %s %s", task.State, task.TaskID)

	switch *update.Status.State {
	case mesosproto.TASK_FAILED, mesosproto.TASK_KILLED, mesosproto.TASK_LOST, mesosproto.TASK_ERROR, mesosproto.TASK_FINISHED:
		e.Redis.DelRedisKey(task.TaskName + ":" + task.TaskID)
		e.API.CleanupNodes()
		return e.Mesos.Call(msg)
	case mesosproto.TASK_RUNNING:
		// remember information for the boostrap server to reach it later
		if task.TaskName == e.Framework.FrameworkName+":server" {
			e.Config.K3SServerContainerPort = int(task.DockerPortMappings[0].HostPort)
			e.Config.K3SServerPort = int(task.DockerPortMappings[1].HostPort)
		}

		task.MesosAgent = e.Mesos.GetAgentInfo(update.Status.GetAgentID().Value)
		task.NetworkInfo = e.Mesos.GetNetworkInfo(task.TaskID)
		task.Agent = update.Status.GetAgentID().Value
		e.Mesos.SuppressFramework()
	}

	e.Redis.SaveTaskRedis(task)
	return e.Mesos.Call(msg)
}
