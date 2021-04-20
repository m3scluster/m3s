package main

import (
	"os"
	"strconv"
	"strings"

	util "github.com/AVENTER-UG/util"

	cfg "mesos-k3s/types"
)

var config cfg.Config

func init() {
	config.K3SAgentMax = 0
	config.K3SServerMax = 0
	config.K3SAgentCount = 0
	config.K3SServerCount = 0

	config.FrameworkUser = util.Getenv("FRAMEWORK_USER", "root")
	config.FrameworkName = util.Getenv("FRAMEWORK_NAME", "k3s")
	config.FrameworkRole = util.Getenv("FRAMEWORK_ROLE", "k3s")
	config.FrameworkPort = util.Getenv("FRAMEWORK_PORT", "10000")
	config.FrameworkInfoFilePath = util.Getenv("FRAMEWORK_STATEFILE_PATH", "/tmp")
	config.Principal = os.Getenv("MESOS_PRINCIPAL")
	config.Username = os.Getenv("MESOS_USERNAME")
	config.Password = os.Getenv("MESOS_PASSWORD")
	config.MesosMasterServer = os.Getenv("MESOS_MASTER")
	config.MesosCNI = util.Getenv("MESOS_CNI", "weave")
	config.LogLevel = util.Getenv("LOGLEVEL", "info")
	config.Domain = os.Getenv("DOMAIN")
	config.K3SAgentMax, _ = strconv.Atoi(util.Getenv("K3S_AGENT_COUNT", "1"))
	config.K3SServerMax, _ = strconv.Atoi(util.Getenv("K3S_SERVER_COUNT", "1"))
	config.ResCPU, _ = strconv.ParseFloat(util.Getenv("RES_CPU", "0.1"), 64)
	config.ResMEM, _ = strconv.ParseFloat(util.Getenv("RES_MEM", "1200"), 64)
	config.Credentials.Username = os.Getenv("AUTH_USERNAME")
	config.Credentials.Password = os.Getenv("AUTH_PASSWORD")
	config.AppName = "Mesos K3S Framework"
	config.K3SCustomDomain = os.Getenv("K3S_CUSTOM_DOMAIN")
	config.K3SServerString = util.Getenv("K3S_SERVER_STRING", "--cluster-cidr=\"10.2.0.0/16\" --service-cidr=\"10.3.0.0/16\" --cluster-dns=\"10.3.0.10\" --disable=metrics-server --snapshotter=native --flannel-backend=vxlan --flannel-iface=ethwe ")
	config.K3SAgentString = util.Getenv("K3S_AGENT_STRING", "--snapshotter=native --flannel-iface=ethwe --flannel-backend=vxlan")
	config.ImageK3S = util.Getenv("IMAGE_K3S", "docker.io/rancher/k3s:v1.20.0-k3s2")
	config.ImageETCD = util.Getenv("IMAGE_ETCD", "docker.io/bitnami/etcd:latest")
	config.VolumeDriver = util.Getenv("VOLUME_DRIVER", "local")
	config.VolumeK3SServer = util.Getenv("VOLUME_K3S_SERVER", "/data/k3s")
	config.PrefixTaskName = util.Getenv("PREFIX_TASKNAME", "k3s")
	config.PrefixHostname = util.Getenv("PREFIX_HOSTNAME", "k3s")
	config.K3SToken = util.Getenv("K3S_TOKEN", "123456789")
	config.ETCDMax, _ = strconv.Atoi(util.Getenv("ETCD_COUNT", "1"))
	config.DockerSock = os.Getenv("DOCKER_SOCK")

	// The comunication to the mesos server should be via ssl or not
	if strings.Compare(os.Getenv("MESOS_SSL"), "true") == 0 {
		config.MesosSSL = true
	} else {
		config.MesosSSL = false
	}

}
