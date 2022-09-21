package mesos

import (
	"encoding/json"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	"github.com/AVENTER-UG/util/util"
	"github.com/sirupsen/logrus"
)

// default resources of the mesos task
func (e *Scheduler) defaultResources(cmd mesosutil.Command) []mesosproto.Resource {
	CPU := "cpus"
	MEM := "mem"
	PORT := "ports"
	cpu := cmd.CPU
	mem := cmd.Memory

	res := []mesosproto.Resource{
		{
			Name:   CPU,
			Type:   mesosproto.SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{Value: cpu},
		},
		{
			Name:   MEM,
			Type:   mesosproto.SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{Value: mem},
		},
	}

	if cmd.DockerPortMappings != nil {
		for _, p := range cmd.DockerPortMappings {
			port := mesosproto.Resource{
				Name: PORT,
				Type: mesosproto.RANGES.Enum(),
				Ranges: &mesosproto.Value_Ranges{
					Range: []mesosproto.Value_Range{
						{
							Begin: uint64(p.HostPort),
							End:   uint64(p.HostPort),
						},
					},
				},
			}
			res = append(res, port)
		}
	}

	return res
}

// default values of the mesos tasks
func (e *Scheduler) defaultCommand(taskID string) mesosutil.Command {
	var cmd mesosutil.Command

	cmd.TaskID = e.getTaskID(taskID)

	cni := e.Framework.MesosCNI
	if e.Framework.MesosCNI == "" {
		if e.Config.DockerCNI != "bridge" {
			cmd.NetworkMode = "user"
			cni = e.Config.DockerCNI
		}
	}
	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &cni,
	}}

	cmd.ContainerType = "DOCKER"

	return cmd
}

func (e *Scheduler) prepareTaskInfoExecuteContainer(agent mesosproto.AgentID, cmd mesosutil.Command) []mesosproto.TaskInfo {
	d, _ := json.Marshal(&cmd)
	logrus.Debug("HandleOffers cmd: ", util.PrettyJSON(d))

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

	var msg mesosproto.TaskInfo

	msg.Name = cmd.TaskName
	msg.TaskID = mesosproto.TaskID{
		Value: cmd.TaskID,
	}
	msg.AgentID = agent
	msg.Resources = e.defaultResources(cmd)

	if cmd.Command == "" {
		msg.Command = &mesosproto.CommandInfo{
			Shell:       &cmd.Shell,
			Arguments:   cmd.Arguments,
			URIs:        cmd.Uris,
			Environment: &cmd.Environment,
		}
	} else {
		msg.Command = &mesosproto.CommandInfo{
			Shell:       &cmd.Shell,
			Value:       &cmd.Command,
			Arguments:   cmd.Arguments,
			URIs:        cmd.Uris,
			Environment: &cmd.Environment,
		}
	}

	msg.Container = &mesosproto.ContainerInfo{
		Type:     contype,
		Volumes:  cmd.Volumes,
		Hostname: &cmd.Hostname,
		Docker: &mesosproto.ContainerInfo_DockerInfo{
			Image:          cmd.ContainerImage,
			Network:        networkMode,
			PortMappings:   cmd.DockerPortMappings,
			Privileged:     &cmd.Privileged,
			Parameters:     cmd.DockerParameter,
			ForcePullImage: func() *bool { x := true; return &x }(),
		},
		NetworkInfos: cmd.NetworkInfo,
	}

	if cmd.Discovery != (mesosproto.DiscoveryInfo{}) {
		msg.Discovery = &cmd.Discovery
	}

	if cmd.Labels != nil {
		msg.Labels = &mesosproto.Labels{
			Labels: cmd.Labels,
		}
	}

	d, _ = json.Marshal(&msg)
	logrus.Debug("HandleOffers msg: ", util.PrettyJSON(d))

	return []mesosproto.TaskInfo{msg}
}
