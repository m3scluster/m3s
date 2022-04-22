package mesos

import (
	"encoding/json"

	api "github.com/AVENTER-UG/mesos-m3s/api"
	mesosutil "github.com/AVENTER-UG/mesos-util"

	mesosproto "github.com/AVENTER-UG/mesos-util/proto"

	"github.com/sirupsen/logrus"
)

// HandleUpdate will handle the offers event of mesos
func HandleUpdate(event *mesosproto.Event) error {
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

	// get the task of the current event, change the state
	task := api.GetTaskFromEvent(update)
	task.State = update.Status.State.String()
	task.Agent = update.Status.GetAgentID().Value

	if task.TaskID == "" {
		return nil
	}

	logrus.Debug(task.State)

	switch *update.Status.State {
	case mesosproto.TASK_FAILED:
		// restart task
		task.State = ""
	case mesosproto.TASK_KILLED:
		// remove task
		api.DelRedisKey(task.TaskName + ":" + task.TaskID)
		return mesosutil.Call(msg)
	case mesosproto.TASK_LOST:
		// restart task
		task.State = ""
	case mesosproto.TASK_ERROR:
		// restart task
		task.State = ""
	}

	// save the new state
	data, _ := json.Marshal(task)
	err := config.RedisClient.Set(config.RedisCTX, task.TaskName+":"+task.TaskID, data, 0).Err()
	if err != nil {
		logrus.Error("HandleUpdate Redis set Error: ", err)
	}

	return mesosutil.Call(msg)
}
