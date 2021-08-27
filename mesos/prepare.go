package mesos

import (
	"strconv"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
)

func prepareTaskInfoExecuteContainer(agent mesosproto.AgentID, cmd cfg.Command) ([]mesosproto.TaskInfo, error) {
	contype := mesosproto.ContainerInfo_DOCKER.Enum()

	// Set Container Network Mode
	networkMode := mesosproto.ContainerInfo_DockerInfo_BRIDGE.Enum()

	if cmd.NetworkMode == "host" {
		networkMode = mesosproto.ContainerInfo_DockerInfo_HOST.Enum()
	}
	if cmd.NetworkMode == "none" {
		networkMode = mesosproto.ContainerInfo_DockerInfo_NONE.Enum()
	}
	if cmd.NetworkMode == "user" {
		networkMode = mesosproto.ContainerInfo_DockerInfo_USER.Enum()
	}
	if cmd.NetworkMode == "bridge" {
		networkMode = mesosproto.ContainerInfo_DockerInfo_BRIDGE.Enum()
	}

	// Save state of the new task
	newTaskID := "k3sframework_" + cmd.TaskName + "_" + strconv.Itoa(int(cmd.TaskID))
	tmp := config.State[newTaskID]
	tmp.Command = cmd
	config.State[newTaskID] = tmp

	var msg mesosproto.TaskInfo

	msg.Name = cmd.TaskName
	msg.TaskID = mesosproto.TaskID{
		Value: newTaskID,
	}
	msg.AgentID = agent
	msg.Resources = defaultResources(cmd)

	msg.Command = &mesosproto.CommandInfo{
		Shell:       &cmd.Shell,
		Value:       &cmd.Command,
		URIs:        cmd.Uris,
		Environment: &cmd.Environment,
	}

	msg.Container = &mesosproto.ContainerInfo{
		Type:     contype,
		Volumes:  cmd.Volumes,
		Hostname: &cmd.Hostname,
		Docker: &mesosproto.ContainerInfo_DockerInfo{
			Image:        cmd.ContainerImage,
			Network:      networkMode,
			PortMappings: cmd.DockerPortMappings,
			Privileged:   &cmd.Privileged,
			Parameters:   cmd.DockerParameter,
		},
		NetworkInfos: cmd.NetworkInfo,
	}

	if cmd.Discovery != (mesosproto.DiscoveryInfo{}) {
		msg.Discovery = &cmd.Discovery
	}

	return []mesosproto.TaskInfo{msg}, nil
}
