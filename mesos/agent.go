package mesos

import (
	"strings"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"

	"github.com/sirupsen/logrus"
)

// StartK3SAgent is starting a agent container with the given IDs
func (e *Scheduler) StartK3SAgent(taskID string) {
	if e.API.CountRedisKey(e.Framework.FrameworkName+":agent:*") >= e.Config.K3SAgentMax {
		return
	}

	cmd := e.defaultCommand(taskID)

	protocol := "tcp"
	cmd.ContainerImage = e.Config.ImageK3S
	cmd.DockerPortMappings = []mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      0,
			ContainerPort: 80,
			Protocol:      &protocol,
		},
		{
			HostPort:      0,
			ContainerPort: 443,
			Protocol:      &protocol,
		},
	}

	cmd.Shell = true
	cmd.Privileged = true
	cmd.Memory = e.Config.K3SAgentMEM
	cmd.CPU = e.Config.K3SAgentCPU
	cmd.TaskName = e.Framework.FrameworkName + ":agent"
	cmd.Hostname = e.Framework.FrameworkName + "agent" + e.Config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + e.Config.K3SAgentString + e.Config.K3SDocker + " --with-node-id " + cmd.TaskID + "'"
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "NET_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "SYS_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "shm-size", Value: e.Config.DockerSHMSize})
	cmd.Instances = e.Config.K3SAgentMax

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

	if e.Config.CGroupV2 {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "cgroupns", Value: "host"})

		cmd.Volumes = []mesosproto.Volume{
			{
				ContainerPath: "/sys/fs/cgroup",
				Mode:          mesosproto.RW.Enum(),
				Source: &mesosproto.Volume_Source{
					Type: mesosproto.Volume_Source_DOCKER_VOLUME,
					DockerVolume: &mesosproto.Volume_Source_DockerVolume{
						Driver: &e.Config.VolumeDriver,
						Name:   func() string { x := "/sys/fs/cgroup"; return x }(),
					},
				},
			},
		}
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

// healthCheckAgent check the health of all agents. Return true if all are fine.
func (e *Scheduler) healthCheckAgent() bool {
	// Hold the at all state of the agent service.
	aState := true

	if e.API.CountRedisKey(e.Framework.FrameworkName+":agent:*") < e.Config.K3SAgentMax {
		return false
	}

	keys := e.API.GetAllRedisKeys(e.Framework.FrameworkName + ":agent:*")
	for keys.Next(e.API.Redis.RedisCTX) {
		key := e.API.GetRedisKey(keys.Val())
		task := mesosutil.DecodeTask(key)

		if task.State == "TASK_RUNNING" {
			aState = aState && true
			continue
		}
		aState = aState && false
	}

	logrus.WithField("func", "healthCheckAgent").Debug("K3s Agent Health: ", aState)
	return aState
}
