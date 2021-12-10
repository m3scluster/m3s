package mesos

import (
	"encoding/json"
	"strings"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	"github.com/AVENTER-UG/util"

	"github.com/sirupsen/logrus"
)

// StartK3SAgent is starting a agent container with the given IDs
func StartK3SAgent(taskID string) {
	var cmd mesosutil.Command

	// if taskID is 0, then its a new task and we have to create a new ID
	newTaskID := taskID
	if taskID == "" {
		newTaskID, _ = util.GenUUID()
	}

	hostport := uint32(getRandomHostPort())
	protocol := "tcp"

	cmd.TaskID = newTaskID

	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = config.ImageK3S
	cmd.NetworkMode = "bridge"
	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &framework.MesosCNI,
	}}

	cmd.DockerPortMappings = []mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      hostport,
			ContainerPort: 80,
			Protocol:      &protocol,
		},
		{
			HostPort:      uint32(hostport + 1),
			ContainerPort: 443,
			Protocol:      &protocol,
		},
	}

	cmd.Shell = true
	cmd.Privileged = true
	cmd.Memory = config.K3SMEM
	cmd.CPU = config.K3SCPU
	cmd.TaskName = config.PrefixHostname + "agent"
	cmd.Hostname = config.PrefixHostname + "agent" + "." + config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + config.K3SAgentString + " --with-node-id " + newTaskID + "'"
	cmd.DockerParameter = []mesosproto.Parameter{
		{
			Key:   "cap-add",
			Value: "NET_ADMIN",
		},
	}
	cmd.Uris = []mesosproto.CommandInfo_URI{
		{
			Value:      config.BootstrapURL,
			Extract:    func() *bool { x := false; return &x }(),
			Executable: func() *bool { x := true; return &x }(),
			Cache:      func() *bool { x := false; return &x }(),
			OutputFile: func() *string { x := "bootstrap"; return &x }(),
		},
	}
	cmd.Volumes = []mesosproto.Volume{
		{
			ContainerPath: "/opt/cni/net.d",
			Mode:          mesosproto.RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME,
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Name: "/etc/mesos/cni/net.d",
				},
			},
		},
	}
	if config.DockerSock != "" {
		cmd.Volumes = []mesosproto.Volume{
			{
				ContainerPath: "/var/run/docker.sock",
				Mode:          mesosproto.RW.Enum(),
				Source: &mesosproto.Volume_Source{
					Type: mesosproto.Volume_Source_DOCKER_VOLUME,
					DockerVolume: &mesosproto.Volume_Source_DockerVolume{
						Name: config.DockerSock,
					},
				},
			},
			{
				ContainerPath: "/opt/cni/net.d",
				Mode:          mesosproto.RW.Enum(),
				Source: &mesosproto.Volume_Source{
					Type: mesosproto.Volume_Source_DOCKER_VOLUME,
					DockerVolume: &mesosproto.Volume_Source_DockerVolume{
						Name: "/etc/mesos/cni/net.d",
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
					Name:     func() *string { x := strings.ToLower(framework.FrameworkName) + "-http"; return &x }(),
					Protocol: cmd.DockerPortMappings[0].Protocol,
				},
				{
					Number:   cmd.DockerPortMappings[1].HostPort,
					Name:     func() *string { x := strings.ToLower(framework.FrameworkName) + "-https"; return &x }(),
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
			Value: &config.K3SToken,
		},
		{
			Name:  "K3S_URL",
			Value: &config.K3SServerURL,
		},
		{
			Name:  "MESOS_SANDBOX_VAR",
			Value: &config.MesosSandboxVar,
		},
	}

	if config.K3SAgentLabels != nil {
		cmd.Labels = config.K3SAgentLabels
	}

	if config.K3SAgentLabels != nil {
		cmd.Labels = config.K3SAgentLabels
	}
	// store mesos task in DB
	d, _ := json.Marshal(&cmd)
	logrus.Debug("Scheduled K3S Agent: ", string(d))
	logrus.Info("Scheduled K3S Agent")
	err := config.RedisClient.Set(config.RedisCTX, cmd.TaskName+":"+newTaskID, d, 0).Err()
	if err != nil {
		logrus.Error("Cloud not store Mesos Task in Redis: ", err)
	}
}
