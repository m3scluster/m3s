package mesos

import (
	"encoding/json"
	"strconv"
	"sync/atomic"

	mesosproto "mesos-k3s/proto"

	cfg "mesos-k3s/types"

	"github.com/sirupsen/logrus"
)

// SearchMissingK3SAgent Check if all zookeepers are running. If one is missing, restart it.
func SearchMissingK3SAgent() {
	if config.State != nil {
		for i := 0; i < config.K3SAgentMax; i++ {
			state := StatusK3SAgent(i)
			if state != nil {
				if *state.Status.State != mesosproto.TASK_RUNNING {
					logrus.Debug("Missing K3SAgent: ", i)
					StartK3SAgent(i)
				}
			}
		}
	}
}

// StatusK3SAgent Get out Status of the given zookeeper ID
func StatusK3SAgent(id int) *cfg.State {
	if config.State != nil {
		for _, element := range config.State {
			if element.Status != nil {
				if element.Command.InternalID == id {
					return &element
				}
			}
		}
	}
	return nil
}

// StartK3SAgent is starting a zookeeper container with the given IDs
func StartK3SAgent(id int) {
	newTaskID := atomic.AddUint64(&config.TaskID, 1)

	var cmd cfg.Command

	// before we will start a new zookeeper, we should be sure its not already running
	status := StatusK3SAgent(id)
	if status != nil {
		if status.Status.State == mesosproto.TASK_STAGING.Enum() {
			logrus.Info("startK3SAgent: kubernetes is staging ", id)
			return
		}
		if status.Status.State == mesosproto.TASK_STARTING.Enum() {
			logrus.Info("startK3SAgent: kubernetes is starting ", id)
			return
		}
		if status.Status.State == mesosproto.TASK_RUNNING.Enum() {
			logrus.Info("startK3SAgent: kubernetes already running ", id)
			return
		}
	}

	var hostport, containerport uint32
	hostport = 31210 + uint32(newTaskID)
	containerport = 2181
	protocol := "tcp"

	cmd.TaskID = newTaskID

	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = config.ImageK3S
	cmd.NetworkMode = "bridge"
	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &config.MesosCNI,
	}}
	cmd.DockerPortMappings = []mesosproto.ContainerInfo_DockerInfo_PortMapping{{
		HostPort:      hostport,
		ContainerPort: containerport,
		Protocol:      &protocol,
	}}

	cmd.Shell = true
	cmd.Privileged = true
	cmd.TaskName = config.PrefixTaskName + "agent" + strconv.Itoa(id)
	cmd.Hostname = config.PrefixTaskName + "agent" + strconv.Itoa(id) + config.K3SCustomDomain + "." + config.Domain
	cmd.Command = "/bin/k3s agent --docker --kubelet-args cgroup-driver=systemd --node-external-ip " + cmd.Hostname + " --with-node-id " + strconv.Itoa(id) + " " + config.K3SAgentString
	cmd.InternalID = id
	cmd.IsK3SAgent = true
	cmd.Volumes = []mesosproto.Volume{
		{
			ContainerPath: "/var/run/docker.sock",
			Mode:          mesosproto.RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME,
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Name: "/var/run/docker.sock",
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
			Name:  "K3S_TOKEN",
			Value: &config.K3SToken,
		},
		{
			Name:  "K3S_URL",
			Value: &config.K3SServerURL,
		},
	}

	d, _ := json.Marshal(&cmd)
	logrus.Debug("Scheduled K3SAgent: ", string(d))

	config.CommandChan <- cmd
	logrus.Info("Scheduled K3SAgent")
}

// The first run have to be in a right sequence
func initStartK3SAgent() {
	if config.K3SAgentCount <= (config.K3SAgentMax - 1) {
		StartK3SAgent(config.K3SAgentCount)
		config.K3SAgentCount++
	}
}

// CreateK3SServerString create the K3S_URL string
func CreateK3SServerString() {
	server := "https://" + config.PrefixHostname + "server" + config.K3SCustomDomain + "." + config.Domain + ":6443"

	config.K3SServerURL = server
}
