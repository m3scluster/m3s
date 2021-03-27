package mesos

import (
	"encoding/json"
	"sync/atomic"

	mesosproto "mesos-k3s/proto"
	cfg "mesos-k3s/types"

	"github.com/sirupsen/logrus"
)

// SearchMissingK3S Check if all k3ss are running. If one is missing, restart it.
func SearchMissingK3SServer() {
	if config.State != nil {
		for i := 0; i < config.K3SServerMax; i++ {
			state := *StatusK3SServer(i).Status.State
			if state != mesosproto.TASK_RUNNING {
				logrus.Debug("Missing K3S: ", i)
				CreateK3SServerString()
				StartK3SServer(i)
			}
		}
	}
}

// StatusK3S Get out Status of the given k3s ID
func StatusK3SServer(id int) *cfg.State {
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

// Start K3S with the given id
func StartK3SServer(id int) {
	newTaskID := atomic.AddUint64(&config.TaskID, 1)

	var cmd cfg.Command

	// be sure, that there is no k3s with this id already running
	status := StatusK3SServer(id)
	if status != nil {
		if status.Status.State == mesosproto.TASK_STAGING.Enum() {
			logrus.Info("startK3SServer: k3s is staging ", id)
			return
		}
		if status.Status.State == mesosproto.TASK_STARTING.Enum() {
			logrus.Info("startK3SServer: k3s is starting ", id)
			return
		}
		if status.Status.State == mesosproto.TASK_RUNNING.Enum() {
			logrus.Info("startK3SServer: k3s already running ", id)
			return
		}
	}
	networkIsolator := "weave"

	cmd.TaskID = newTaskID
	cmd.Command = "/bin/k3s server --cluster-cidr \"10.1.0.0/16\" --service-cidr \"10.2.0.0/16\" --cluster-dns 10.2.0.10 "
	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = config.ImageK3S
	cmd.NetworkMode = "bridge"
	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &networkIsolator,
	}}
	cmd.Shell = true
	cmd.Privileged = true
	cmd.InternalID = id
	cmd.IsK3SServer = true
	cmd.TaskName = config.PrefixTaskName + "server"
	cmd.Hostname = config.PrefixHostname + "server"
	cmd.Domain = config.K3SCustomString + "." + config.Domain
	cmd.Volumes = []mesosproto.Volume{
		{
			ContainerPath: "/var/lib/rancher/k3s",
			Mode:          mesosproto.RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME,
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Driver: &config.VolumeDriver,
					Name:   config.VolumeK3SServer,
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
			Name:  "K3S_KUBECONFIG_OUTPUT",
			Value: func() *string { x := "/output/kubeconfig.yaml"; return &x }(),
		},
		{
			Name:  "K3S_KUBECONFIG_MODE",
			Value: func() *string { x := "666"; return &x }(),
		},
	}

	d, _ := json.Marshal(&cmd)
	logrus.Debug("Scheduled K3S: ", string(d))

	config.CommandChan <- cmd
	logrus.Info("Scheduled K3S")

}

// the first run should be in ta strict order.
func initStartK3SServer() {
	if config.K3SServerCount <= (config.K3SServerMax - 1) {
		StartK3SServer(config.K3SServerCount)
		config.K3SServerCount++
	}
}
