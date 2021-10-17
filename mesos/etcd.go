package mesos

import (
	"encoding/json"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"

	"github.com/sirupsen/logrus"
)

// SearchMissingEtcd Check if all agents are running. If one is missing, restart it.
func SearchMissingEtcd(restart bool) bool {
	status := make([]mesosproto.TaskState, config.ETCDMax)
	allRunning := true

	if config.State != nil {
		for i := 0; i < config.ETCDMax; i++ {
			state := StatusEtcd(i)
			if state != nil {
				status[i] = *state.Status.State
				if *state.Status.State != mesosproto.TASK_RUNNING {
					allRunning = false
					logrus.Debug("Missing ETCD: ", i)
					if restart {
						StartEtcd(i)
					}
				}
			} else {
				allRunning = false
			}
		}
	} else {
		allRunning = false
	}
	config.M3SStatus.Etcd = status
	return allRunning
}

// StatusEtcd Get out Status of the given ID
func StatusEtcd(id int) *cfg.State {
	if config.State != nil {
		for _, element := range config.State {
			if element.Status != nil {
				if element.Command.InternalID == id && element.Command.IsETCD {
					return &element
				}
			}
		}
	}
	return nil
}

// StartEtcd is starting the etcd
func StartEtcd(id int) {
	var cmd cfg.Command

	// before we will start a new agent, we should be sure its not already running
	status := StatusEtcd(id)
	if status != nil {
		if status.Status.State == mesosproto.TASK_STAGING.Enum() {
			logrus.Info("startETCD: etcd is staging ", id)
			return
		}
		if status.Status.State == mesosproto.TASK_STARTING.Enum() {
			logrus.Info("startETCD: etcd is starting ", id)
			return
		}
		if status.Status.State == mesosproto.TASK_RUNNING.Enum() {
			logrus.Info("startETCD: etcd already running ", id)
			return
		}
	}

	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = config.ImageETCD
	cmd.NetworkMode = "bridge"

	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &config.MesosCNI,
	}}
	cmd.Shell = true
	cmd.Privileged = false
	cmd.InternalID = id
	cmd.Memory = config.ETCDMEM
	cmd.CPU = config.ETCDCPU
	cmd.TaskName = config.PrefixTaskName + "etcd"
	cmd.Hostname = config.PrefixTaskName + "etcd" + "." + config.Domain
	cmd.IsETCD = true
	cmd.DockerParameter = []mesosproto.Parameter{}

	AllowNoneAuthentication := "yes"
	AdvertiseURL := "http://" + cmd.Hostname + ":2379"

	cmd.Command = "/opt/bitnami/etcd/bin/etcd --listen-client-urls http://0.0.0.0:2379"

	cmd.Environment.Variables = []mesosproto.Environment_Variable{
		{
			Name:  "SERVICE_NAME",
			Value: &cmd.TaskName,
		},
		{
			Name:  "ALLOW_NONE_AUTHENTICATION",
			Value: &AllowNoneAuthentication,
		},
		{
			Name:  "ETCD_ADVERTISE_CLIENT_URLS",
			Value: &AdvertiseURL,
		},
		{
			Name: "ETCD_DATA_DIR",
			Value: func() *string {
				x := "/tmp"
				return &x
			}(),
		},
	}

	cmd.Discovery = mesosproto.DiscoveryInfo{
		Visibility: 1,
		Name:       &cmd.TaskName,
	}

	d, _ := json.Marshal(&cmd)
	logrus.Debug("Scheduled Etcd: ", string(d))

	config.CommandChan <- cmd
	logrus.Info("Scheduled Etcd")
}

// The first run have to be in a right sequence
func initStartEtcd() {
	if config.ETCDCount <= (config.ETCDMax - 1) {
		StartEtcd(config.ETCDCount)
		config.ETCDCount++
	}
}
