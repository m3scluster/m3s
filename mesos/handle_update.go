package mesos

import (
	"encoding/json"
	"io/ioutil"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"

	"github.com/sirupsen/logrus"
)

// HandleUpdate will handle the offers event of mesos
func HandleUpdate(event *mesosproto.Event) error {
	// unsuppress
	revive := &mesosproto.Call{
		Type: mesosproto.Call_REVIVE,
	}
	Call(revive)

	update := event.Update

	msg := &mesosproto.Call{
		Type: mesosproto.Call_ACKNOWLEDGE,
		Acknowledge: &mesosproto.Call_Acknowledge{
			AgentID: *update.Status.AgentID,
			TaskID:  update.Status.TaskID,
			UUID:    update.Status.UUID,
		},
	}

	// Save state of the task
	taskID := update.Status.GetTaskID().Value
	tmp := config.State[taskID]
	tmp.Status = &update.Status

	logrus.Debugf("HandleUpdate: %s Status %s ", taskID, update.Status.State.String())

	switch *update.Status.State {
	case mesosproto.TASK_FAILED:
		deleteOldTask(tmp.Status.TaskID)
	case mesosproto.TASK_KILLED:
		deleteOldTask(tmp.Status.TaskID)
	case mesosproto.TASK_LOST:
		deleteOldTask(tmp.Status.TaskID)
	}

	// Update Framework State File
	config.State[taskID] = tmp
	persConf, _ := json.Marshal(&config)
	ioutil.WriteFile(config.FrameworkInfoFile, persConf, 0644)

	return Call(msg)
}
