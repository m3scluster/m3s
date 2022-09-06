package mesos

import (
	mesosutil "github.com/AVENTER-UG/mesos-util"

	mesosproto "github.com/AVENTER-UG/mesos-util/proto"

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
	task := e.API.GetTaskFromEvent(update)

	// if these object have not TaskID it, invalid
	if task.TaskID == "" {
		return nil
	}

	task.State = update.Status.State.String()

	logrus.WithField("func", "HandleUpdate").Debug("Task State: ", task.State)

	switch *update.Status.State {
	case mesosproto.TASK_FAILED, mesosproto.TASK_KILLED, mesosproto.TASK_LOST, mesosproto.TASK_ERROR, mesosproto.TASK_FINISHED:
		e.API.DelRedisKey(task.TaskName + ":" + task.TaskID)
		e.API.CleanupNodes()
		return mesosutil.Call(msg)
	case mesosproto.TASK_RUNNING:
		// remember information for the boostrap server to reach it later
		if task.TaskName == e.Framework.FrameworkName+":server" {
			e.Config.K3SServerContainerPort = int(task.DockerPortMappings[0].HostPort)
			e.Config.K3SServerPort = int(task.DockerPortMappings[1].HostPort)
		}

		task.MesosAgent = mesosutil.GetAgentInfo(update.Status.GetAgentID().Value)
		task.NetworkInfo = mesosutil.GetNetworkInfo(task.TaskID)
		task.Agent = update.Status.GetAgentID().Value
		mesosutil.SuppressFramework()
	}

	e.API.SaveTaskRedis(task)
	return mesosutil.Call(msg)
}
