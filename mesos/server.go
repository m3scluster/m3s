package mesos

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync/atomic"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"

	"github.com/sirupsen/logrus"
)

// SearchMissingK3SServer Check if all k3ss are running. If one is missing, restart it.
func SearchMissingK3SServer() {
	if config.State != nil {
		for i := 0; i < config.K3SServerMax; i++ {
			state := StatusK3SServer(i)
			if state != nil {
				if *state.Status.State != mesosproto.TASK_RUNNING {
					logrus.Debug("Missing K3S: ", i)
					CreateK3SServerString()
					StartK3SServer(i)
				}
			}
		}
	}
}

// StatusK3SServer Get out Status of the given k3s ID
func StatusK3SServer(id int) *cfg.State {
	if config.State != nil {
		for _, element := range config.State {
			if element.Status != nil {
				if element.Command.InternalID == id && element.Command.IsK3SServer == true {
					config.M3SStatus.Server = element.Status.State
					return &element
				}
			}
		}
	}
	config.M3SStatus.Server = mesosproto.TASK_UNKNOWN.Enum()
	return nil
}

// StartK3SServer Start K3S with the given id
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

	cmd.TaskID = newTaskID

	cmd.ContainerType = "DOCKER"
	cmd.ContainerImage = config.ImageK3S
	cmd.NetworkMode = "bridge"
	cmd.NetworkInfo = []mesosproto.NetworkInfo{{
		Name: &config.MesosCNI,
	}}

	cmd.Shell = true
	cmd.Privileged = true
	cmd.InternalID = id
	cmd.IsK3SServer = true
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

	hostport := 31859 + uint32(newTaskID)
	protocol := "tcp"
	cmd.DockerPortMappings = []mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      hostport,
			ContainerPort: 10422,
			Protocol:      &protocol,
		},
		{
			HostPort:      hostport + 1,
			ContainerPort: 6443,
			Protocol:      &protocol,
		},
		{
			HostPort:      hostport + 2,
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
	}

	d, _ := json.Marshal(&cmd)
	logrus.Debug("Scheduled K3S Server: ", string(d))

	config.CommandChan <- cmd
	logrus.Info("Scheduled K3S Server")

}

// the first run should be in ta strict order.
func initStartK3SServer() {
	etcdState := StatusEtcd(config.ETCDMax - 1)

	// if etcd still not run, then do not start K3S Manager
	if etcdState == nil {
		return
	}

	if config.K3SServerCount <= (config.K3SServerMax-1) && etcdState.Status.GetState() == 1 {
		StartK3SServer(config.K3SServerCount)
		config.K3SServerCount++
	}
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

// K3SHeartbeat to execute K3S Bootstrap API Server commands
func K3SHeartbeat() {
	if !IsK3SServerRunning() {
		initStartK3SServer()
	} else {
		initStartK3SAgent()
	}
}
