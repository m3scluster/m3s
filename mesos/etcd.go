package mesos

import (
	"encoding/json"

	api "github.com/AVENTER-UG/mesos-m3s/api"
	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	"github.com/AVENTER-UG/util"

	"github.com/sirupsen/logrus"
)

func getEtcdStatus() string {
	keys := api.GetAllRedisKeys("k3setcd:*")

	for keys.Next(config.RedisCTX) {
		key := api.GetRedisKey(keys.Val())
		var task mesosutil.Command
		json.Unmarshal([]byte(key), &task)
		return task.State
	}
	return ""
}

// StartEtcd is starting the etcd
func StartEtcd(taskID string) {
	var cmd mesosutil.Command

	// if taskID is 0, then its a new task and we have to create a new ID
	newTaskID := taskID
	if taskID == "" {
		newTaskID, _ = util.GenUUID()
	}

	cmd.TaskID = newTaskID
	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = config.ImageETCD
	cmd.NetworkMode = "bridge"

	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &framework.MesosCNI,
	}}
	cmd.Shell = true
	cmd.Privileged = false
	cmd.Memory = config.ETCDMEM
	cmd.CPU = config.ETCDCPU
	cmd.TaskName = "k3setcd"
	cmd.Hostname = "k3setcd" + "." + config.Domain
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

	// store mesos task in DB
	d, _ := json.Marshal(&cmd)
	logrus.Debug("Scheduled Etcd: ", string(d))
	logrus.Info("Scheduled Etcd")
	err := config.RedisClient.Set(config.RedisCTX, cmd.TaskName+":"+newTaskID, d, 0).Err()
	if err != nil {
		logrus.Error("Cloud not store Mesos Task in Redis: ", err)
	}
}
