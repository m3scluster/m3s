package mesos

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	"github.com/AVENTER-UG/util"

	"github.com/sirupsen/logrus"
)

// StartK3SServer Start K3S with the given id
func StartK3SServer(taskID string) {
	// if k3s max count is reach, do not start a new server
	if config.K3SServerCount == config.K3SServerMax {
		return
	}

	// if taskID is 0, then its a new task and we have to create a new ID
	newTaskID := taskID
	if taskID == "" {
		newTaskID, _ = util.GenUUID()
	}

	var cmd mesosutil.Command

	cmd.TaskID = newTaskID

	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = config.ImageK3S
	cmd.NetworkMode = "bridge"
	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &framework.MesosCNI,
	}}

	cmd.Shell = true
	cmd.Privileged = true
	cmd.ContainerImage = config.ImageK3S
	cmd.Memory = config.K3SMEM
	cmd.CPU = config.K3SCPU
	cmd.TaskName = config.PrefixTaskName + "server"
	cmd.Hostname = config.PrefixHostname + "server" + "." + config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + config.K3SServerString + "--tls-san=" + config.Domain + "'"
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

	protocol := "tcp"
	cmd.DockerPortMappings = []mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      uint32(getRandomHostPort()),
			ContainerPort: 10422,
			Protocol:      &protocol,
		},
		{
			HostPort:      uint32(getRandomHostPort()),
			ContainerPort: 6443,
			Protocol:      &protocol,
		},
		{
			HostPort:      uint32(getRandomHostPort()),
			ContainerPort: 8080,
			Protocol:      &protocol,
		},
	}

	cmd.Discovery = mesosproto.DiscoveryInfo{
		Visibility: 2,
		Name:       &cmd.TaskName,
		Ports: &mesosproto.Ports{
			Ports: []mesosproto.Port{
				{
					Number:   cmd.DockerPortMappings[0].HostPort,
					Name:     func() *string { x := "api"; return &x }(),
					Protocol: cmd.DockerPortMappings[0].Protocol,
				},
				{
					Number:   cmd.DockerPortMappings[1].HostPort,
					Name:     func() *string { x := "kubernetes"; return &x }(),
					Protocol: cmd.DockerPortMappings[1].Protocol,
				},
			},
		},
	}

	CreateK3SServerString()

	cmd.Environment.Variables = []mesosproto.Environment_Variable{
		{
			Name:  "SERVICE_NAME",
			Value: &cmd.TaskName,
		},
		{
			Name:  "K3SFRAMEWORK_TYPE",
			Value: func() *string { x := "server"; return &x }(),
		},
		{
			Name:  "K3S_URL",
			Value: &config.K3SServerURL,
		},
		{
			Name:  "K3S_TOKEN",
			Value: &config.K3SToken,
		},
		{
			Name:  "K3S_KUBECONFIG_OUTPUT",
			Value: func() *string { x := "/mnt/mesos/sandbox/kubeconfig.yaml"; return &x }(),
		},
		{
			Name:  "K3S_KUBECONFIG_MODE",
			Value: func() *string { x := "666"; return &x }(),
		},
		{
			Name: "K3S_DATASTORE_ENDPOINT",
			Value: func() *string {
				x := "http://" + config.PrefixTaskName + "etcd" + "." + config.Domain + ":2379"
				return &x
			}(),
		},
		{
			Name:  "MESOS_SANDBOX_VAR",
			Value: &config.MesosSandboxVar,
		},
	}

	// store mesos task in DB
	d, _ := json.Marshal(&cmd)
	logrus.Debug("Scheduled K3S Server: ", string(d))
	logrus.Info("Scheduled K3S Server")
	err := config.RedisClient.Set(config.RedisCTX, cmd.TaskName+":"+newTaskID, d, 0).Err()
	if err != nil {
		logrus.Error("Cloud not store Mesos Task in Redis: ", err)
	}

	config.K3SServerCount = config.K3SServerCount + 1
	logrus.Debug(config.K3SServerCount)

}

// CreateK3SServerString create the K3S_URL string
func CreateK3SServerString() {
	server := "https://" + config.PrefixHostname + "server" + "." + config.Domain + ":6443"

	config.K3SServerURL = server
}

// IsK3SServerRunning check if the kubernetes server is already running
func IsK3SServerRunning() bool {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+config.M3SBootstrapServerHostname+":"+strconv.Itoa(config.M3SBootstrapServerPort)+"/status", nil)
	req.Close = true
	res, err := client.Do(req)

	if err != nil {
		logrus.Error("IsK3SServerRunning: Error 1: ", err, res)
		return false
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.Error("IsK3SServerRunning: Error Status is not 200")
		return false
	}

	content, err := ioutil.ReadAll(res.Body)

	if err != nil {
		logrus.Error("IsK3SServerRunning: Error 2: ", err, res)
		return false
	}

	if string(content) == "ok" {
		logrus.Debug("IsK3SServerRunning: True")
		config.M3SStatus.API = "ok"
		return true
	}

	config.M3SStatus.API = "nok"
	return false
}
