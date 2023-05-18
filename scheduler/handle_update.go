package scheduler

import (
	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"

	"github.com/sirupsen/logrus"
)

// HandleUpdate will handle the offers event of mesos
func (e *Scheduler) HandleUpdate(event *mesosproto.Event) error {
	update := event.Update

	if update.Status.UUID == nil {
		logrus.WithField("func", "scheduler.HandleUpdate").Debug("UUID is not set")
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

	logrus.WithField("func", "scheduler.HandleUpdate").Debugf("Task State: %s %s", task.State, task.TaskID)

	switch *update.Status.State {
	case mesosproto.TASK_FAILED, mesosproto.TASK_KILLED, mesosproto.TASK_LOST, mesosproto.TASK_ERROR, mesosproto.TASK_FINISHED:
		e.Redis.DelRedisKey(task.TaskName + ":" + task.TaskID)
		e.API.CleanupNodes()
		return e.Mesos.Call(msg)
	case mesosproto.TASK_RUNNING:
		task.MesosAgent = e.Mesos.GetAgentInfo(update.Status.GetAgentID().Value)
		task.NetworkInfo = e.Mesos.GetNetworkInfo(task.TaskID)
		task.Agent = update.Status.GetAgentID().Value
		// remember information for the boostrap server to reach it later
		if task.TaskName == e.Framework.FrameworkName+":server" {
			// if the framework is running as container, and the task hostname is the same like the frameworks one,
			// then use the containerport instead of the random hostport
			if e.Config.DockerRunning && (task.MesosAgent.Hostname == e.Config.Hostname) {
				e.Config.K3SServerContainerPort = int(task.DockerPortMappings[0].ContainerPort)
				e.Config.K3SServerHostname = task.Hostname
			} else {
				e.Config.K3SServerContainerPort = int(task.DockerPortMappings[0].HostPort)
			}
			e.Config.K3SServerPort = int(task.DockerPortMappings[1].HostPort)

			e.API.AgentRestart()
		}
	}

	e.Redis.SaveTaskRedis(task)
	return e.Mesos.Call(msg)
}
