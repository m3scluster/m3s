package mesos

import (
	"strings"

	mesosproto "github.com/AVENTER-UG/mesos-util/proto"

	"github.com/sirupsen/logrus"
)

// StartK3SAgent is starting a agent container with the given IDs
func (e *Scheduler) StartK3SAgent(taskID string) {
	cmd := e.defaultCommand(taskID)

	hostport := e.getRandomHostPort(2)
	if hostport == 0 {
		logrus.WithField("func", "StartK3SAgent").Error("Could not find free ports")
		return
	}
	protocol := "tcp"

	cmd.ContainerImage = e.Config.ImageK3S

	cmd.DockerPortMappings = []mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      hostport,
			ContainerPort: 80,
			Protocol:      &protocol,
		},
		{
			HostPort:      hostport + 1,
			ContainerPort: 443,
			Protocol:      &protocol,
		},
	}

	cmd.Privileged = true
	cmd.Memory = e.Config.K3SAgentMEM
	cmd.CPU = e.Config.K3SAgentCPU
	cmd.TaskName = e.Framework.FrameworkName + ":agent"
	cmd.Hostname = e.Framework.FrameworkName + "agent" + e.Config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + e.Config.K3SAgentString + e.Config.K3SDocker + " --with-node-id " + cmd.TaskID + "'"
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "NET_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "SYS_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "shm-size", Value: e.Config.DockerSHMSize})
	// if mesos cni is unset, then use docker cni
	if e.Framework.MesosCNI == "" {
		// net-alias is only supported onuser-defined networks
		if e.Config.DockerCNI != "bridge" {
			cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "net-alias", Value: e.Framework.FrameworkName + "agent"})
		}
	}

	cmd.Uris = []mesosproto.CommandInfo_URI{
		{
			Value:      e.Config.BootstrapURL,
			Extract:    func() *bool { x := false; return &x }(),
			Executable: func() *bool { x := true; return &x }(),
			Cache:      func() *bool { x := false; return &x }(),
			OutputFile: func() *string { x := "bootstrap"; return &x }(),
		},
	}

	cmd.Discovery = mesosproto.DiscoveryInfo{
		Visibility: 2,
		Name:       &cmd.TaskName,
		Ports: &mesosproto.Ports{
			Ports: []mesosproto.Port{
				{
					Number:   cmd.DockerPortMappings[0].HostPort,
					Name:     func() *string { x := strings.ToLower(e.Framework.FrameworkName) + "-http"; return &x }(),
					Protocol: cmd.DockerPortMappings[0].Protocol,
				},
				{
					Number:   cmd.DockerPortMappings[1].HostPort,
					Name:     func() *string { x := strings.ToLower(e.Framework.FrameworkName) + "-https"; return &x }(),
					Protocol: cmd.DockerPortMappings[1].Protocol,
				},
			},
		},
	}

	cmd.Environment.Variables = []mesosproto.Environment_Variable{
		{
			Name:  "SERVICE_NAME",
			Value: &cmd.TaskName,
		},
		{
			Name:  "K3SFRAMEWORK_TYPE",
			Value: func() *string { x := "agent"; return &x }(),
		},
		{
			Name:  "K3S_TOKEN",
			Value: &e.Config.K3SToken,
		},
		{
			Name:  "K3S_URL",
			Value: &e.Config.K3SServerURL,
		},
		{
			Name:  "MESOS_SANDBOX_VAR",
			Value: &e.Config.MesosSandboxVar,
		},
	}

	if e.Config.K3SAgentLabels != nil {
		cmd.Labels = e.Config.K3SAgentLabels
	}

	if e.Config.K3SAgentLabels != nil {
		cmd.Labels = e.Config.K3SAgentLabels
	}

	// store mesos task in DB
	logrus.WithField("func", "StartK3SAgent").Info("Schedule K3S Agent")
	e.API.SaveTaskRedis(cmd)
}
