package mesos

import (
	"io/ioutil"
	"net/http"
	"strconv"

	mesosproto "github.com/AVENTER-UG/mesos-util/proto"

	"github.com/sirupsen/logrus"
)

// StartK3SServer Start K3S with the given id
func (e *Scheduler) StartK3SServer(taskID string) {
	cmd := e.defaultCommand(taskID)

	cmd.Shell = true
	cmd.ContainerImage = e.Config.ImageK3S
	cmd.Privileged = true
	cmd.ContainerImage = e.Config.ImageK3S
	cmd.Memory = e.Config.K3SServerMEM
	cmd.CPU = e.Config.K3SServerCPU
	cmd.TaskName = e.Framework.FrameworkName + ":server"
	cmd.Hostname = e.Framework.FrameworkName + "server" + e.Config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + e.Config.K3SServerString + e.Config.K3SDocker + " --kube-controller-manager-arg='leader-elect=false' --kube-scheduler-arg='leader-elect=false' -tls-san=" + e.Framework.FrameworkName + "server'"
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "NET_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "SYS_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "shm-size", Value: e.Config.DockerSHMSize})
	// if mesos cni is unset, then use docker cni
	if e.Framework.MesosCNI == "" {
		// net-alias is only supported onuser-defined networks
		if e.Config.DockerCNI != "bridge" {
			cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "net-alias", Value: e.Framework.FrameworkName + "server"})
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
	cmd.Volumes = []mesosproto.Volume{
		{
			ContainerPath: "/var/lib/rancher/k3s",
			Mode:          mesosproto.RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME,
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Driver: &e.Config.VolumeDriver,
					Name:   e.Config.VolumeK3SServer,
				},
			},
		},
	}

	// get free hostport. If there is no one, do not start
	hostport := e.getRandomHostPort(3)
	if hostport == 0 {
		logrus.WithField("func", "StartK3SServer").Error("Could not find free ports")
		return
	}
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
				{
					Number:   cmd.DockerPortMappings[2].HostPort,
					Name:     func() *string { x := "http"; return &x }(),
					Protocol: cmd.DockerPortMappings[2].Protocol,
				},
			},
		},
	}

	e.CreateK3SServerString()

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
			Value: &e.Config.K3SServerURL,
		},
		{
			Name:  "K3S_TOKEN",
			Value: &e.Config.K3SToken,
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
			Name:  "MESOS_SANDBOX_VAR",
			Value: &e.Config.MesosSandboxVar,
		},
	}

	if e.Config.DSEtcd {
		ds := mesosproto.Environment_Variable{
			Name: "K3S_DATASTORE_ENDPOINT",
			Value: func() *string {
				x := "http://" + e.Framework.FrameworkName + "datastore" + e.Config.Domain + ":" + e.Config.DSPort
				return &x
			}(),
		}
		cmd.Environment.Variables = append(cmd.Environment.Variables, ds)
	}

	if e.Config.DSMySQL {
		ds := mesosproto.Environment_Variable{
			Name: "K3S_DATASTORE_ENDPOINT",
			Value: func() *string {
				x := "mysql://" + e.Config.DSMySQLUsername + ":" + e.Config.DSMySQLPassword + "@tcp(" + e.Framework.FrameworkName + "datastore" + e.Config.Domain + ":" + e.Config.DSPort + ")/k3s"
				return &x
			}(),
		}
		cmd.Environment.Variables = append(cmd.Environment.Variables, ds)
	}

	// store mesos task in DB
	logrus.WithField("func", "StartK3SServer").Info("Schedule K3S Server")
	e.API.SaveTaskRedis(cmd)
}

// CreateK3SServerString create the K3S_URL string
func (e *Scheduler) CreateK3SServerString() {
	server := "https://" + e.Framework.FrameworkName + "server" + e.Config.Domain + ":6443"

	e.Config.K3SServerURL = server
}

// healthCheckK3s check if the kubernetes server is already running
func (e *Scheduler) healthCheckK3s() bool {
	k3sState := false

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+e.Config.M3SBootstrapServerHostname+":"+strconv.Itoa(e.Config.M3SBootstrapServerPort)+"/api/m3s/bootstrap/v0/status", nil)
	req.Close = true
	res, err := client.Do(req)
	var content []byte

	if err != nil {
		k3sState = false
		goto end
	}

	defer res.Body.Close()

	content, _ = ioutil.ReadAll(res.Body)

	if string(content) == "ok" {
		e.Config.M3SStatus.API = "ok"
		k3sState = true
	} else {
		e.Config.M3SStatus.API = "nok"
		k3sState = false
	}

end:
	logrus.WithField("func", "healthCheckK3s").Debug("K3s Manager Health: ", k3sState)
	return k3sState
}
