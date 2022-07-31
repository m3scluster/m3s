package mesos

import (
	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"

	"github.com/sirupsen/logrus"
)

// StartEtcd is starting the etcd
func (e *Scheduler) StartEtcd(taskID string) {
	cmd := e.defaultCommand(taskID)

	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = e.Config.ImageETCD
	cmd.Privileged = false
	cmd.Memory = e.Config.ETCDMEM
	cmd.CPU = e.Config.ETCDCPU
	cmd.Disk = e.Config.ETCDDISK
	cmd.TaskName = e.Framework.FrameworkName + ":etcd"
	cmd.Hostname = e.Framework.FrameworkName + "etcd" + e.Config.Domain
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "NET_ADMIN"})
	// if mesos cni is unset, then use docker cni
	if e.Framework.MesosCNI == "" {
		// net-alias is only supported onuser-defined networks
		if e.Config.DockerCNI != "bridge" {
			cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "net-alias", Value: e.Framework.FrameworkName + "etcd"})
		}
	}

	AllowNoneAuthentication := "yes"
	AdvertiseURL := "http://" + cmd.Hostname + ":2379"

	cmd.Command = "/opt/bitnami/etcd/bin/etcd --listen-client-urls http://0.0.0.0:2379 --election-timeout '50000' --heartbeat-interval '5000'"

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

	// store mesos task in DB
	logrus.WithField("func", "StartETCD").Info("Schedule ETCD")
	e.API.SaveTaskRedis(cmd)
}

func (e *Scheduler) addDockerParameter(current []mesosproto.Parameter, newValues mesosproto.Parameter) []mesosproto.Parameter {
	return append(current, newValues)
}

// healthCheckETCD check the health of all etcdservers. Return true if all are fine.
func (e *Scheduler) healthCheckETCD() bool {
	// Hold the at all state of the etcd service.
	etcdState := false

	keys := e.API.GetAllRedisKeys(e.Framework.FrameworkName + ":etcd:*")
	for keys.Next(e.API.Redis.RedisCTX) {
		key := e.API.GetRedisKey(keys.Val())
		task := mesosutil.DecodeTask(key)

		if task.State == "TASK_RUNNING" && len(task.NetworkInfo) > 0 {
			etcdState = true
		}
	}

	logrus.WithField("func", "healthCheckETCD").Debug("ETCD Health: ", etcdState)
	return etcdState
}
