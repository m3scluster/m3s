package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	util "github.com/AVENTER-UG/util/util"
	"github.com/Showmax/go-fqdn"
	"github.com/sirupsen/logrus"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"
)

var config cfg.Config
var framework mesosutil.FrameworkConfig

func init() {
	framework.FrameworkUser = util.Getenv("FRAMEWORK_USER", "root")
	framework.FrameworkName = "m3s" + util.Getenv("FRAMEWORK_NAME", "")
	framework.FrameworkRole = util.Getenv("FRAMEWORK_ROLE", "m3s")
	framework.FrameworkPort = util.Getenv("FRAMEWORK_PORT", "10000")
	framework.FrameworkHostname = util.Getenv("FRAMEWORK_HOSTNAME", fqdn.Get())
	framework.FrameworkInfoFilePath = util.Getenv("FRAMEWORK_STATEFILE_PATH", "/tmp")
	framework.Username = util.Getenv("MESOS_USERNAME", "")
	framework.Password = util.Getenv("MESOS_PASSWORD", "")
	framework.MesosMasterServer = util.Getenv("MESOS_MASTER", "127.0.0.1:5050")
	framework.MesosCNI = os.Getenv("MESOS_CNI")
	framework.MesosSSL = stringToBool(os.Getenv("MESOS_SSL"))
	framework.PortRangeFrom, _ = strconv.Atoi(util.Getenv("PORTRANGE_FROM", "31000"))
	framework.PortRangeTo, _ = strconv.Atoi(util.Getenv("PORTRANGE_TO", "32000"))
	config.AppName = "Mesos K3S Framework"
	config.BootstrapURL = util.Getenv("BOOTSTRAP_URL", "https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/dev/bootstrap/bootstrap.sh")
	config.BootstrapCredentials.Username = util.Getenv("BOOTSTRAP_AUTH_USERNAME", "")
	config.BootstrapCredentials.Password = util.Getenv("BOOTSTRAP_AUTH_PASSWORD", "")
	config.BootstrapSSLKey = util.Getenv("BOOTSTRAP_SSL_KEY_BASE64", "")
	config.BootstrapSSLCrt = util.Getenv("BOOTSTRAP_SSL_CRT_BASE64", "")
	config.CGroupV2 = stringToBool(util.Getenv("CGROUP_V2", "false"))
	config.DSMax, _ = strconv.Atoi(util.Getenv("DS_COUNT", "1"))
	config.EventLoopTime, _ = time.ParseDuration(util.Getenv("HEARTBEAT_INTERVAL", "15s"))
	config.CleanupLoopTime, _ = time.ParseDuration(util.Getenv("CLEANUP_WAIT", "5m"))
	config.ReviveLoopTime, _ = time.ParseDuration(util.Getenv("REVIVE_WAIT", "1m"))
	config.Credentials.Username = util.Getenv("AUTH_USERNAME", "")
	config.Credentials.Password = util.Getenv("AUTH_PASSWORD", "")
	config.Domain = util.Getenv("DOMAIN", ".local")
	config.DockerSock = os.Getenv("DOCKER_SOCK")
	config.DockerSHMSize = util.Getenv("DOCKER_SHM_SIZE", "30gb")
	config.DockerCNI = util.Getenv("DOCKER_CNI", "bridge")
	config.DSCPU, _ = strconv.ParseFloat(util.Getenv("DS_CPU", "0.1"), 64)
	config.DSMEM, _ = strconv.ParseFloat(util.Getenv("DS_MEM", "1000"), 64)
	config.DSDISK, _ = strconv.ParseFloat(util.Getenv("DS_DISK", "10000"), 64)
	config.DSEtcd = stringToBool(os.Getenv("DS_ETCD"))
	config.DSPort = util.Getenv("DS_PORT", "3306")
	config.DSMySQL = stringToBool(util.Getenv("DS_MYSQL", "true"))
	config.DSMySQLUsername = util.Getenv("DS_MYSQL_USERNAME", "root")
	config.DSMySQLPassword = util.Getenv("DS_MYSQL_PASSWORD", "password")
	config.DSMySQLSSL = stringToBool(util.Getenv("DS_MYSQL_SSL", "false"))
	config.K3SServerMax, _ = strconv.Atoi(util.Getenv("K3S_SERVER_COUNT", "1"))
	config.K3SServerContainerPort, _ = strconv.Atoi(util.Getenv("K3S_SERVER_PORT", "6443"))
	config.K3SServerCPU, _ = strconv.ParseFloat(util.Getenv("K3S_SERVER_CPU", "0.1"), 64)
	config.K3SServerMEM, _ = strconv.ParseFloat(util.Getenv("K3S_SERVER_MEM", "2000"), 64)
	config.K3SServerString = util.Getenv("K3S_SERVER_STRING", "/usr/local/bin/k3s server --cluster-cidr=10.2.0.0/16 --service-cidr=10.3.0.0/16 --cluster-dns=10.3.0.10  --kube-controller-manager-arg='leader-elect=false' --disable-cloud-controller --kube-scheduler-arg='leader-elect=false' --snapshotter=native --flannel-backend=vxlan ")
	config.K3SCustomDomain = util.Getenv("K3S_CUSTOM_DOMAIN", "cloud.local")
	config.K3SAgentString = util.Getenv("K3S_AGENT_STRING", "/usr/local/bin/k3s agent --snapshotter=native --flannel-backend=vxlan ")
	config.K3SAgentMax, _ = strconv.Atoi(util.Getenv("K3S_AGENT_COUNT", "1"))
	config.K3SAgentCPU, _ = strconv.ParseFloat(util.Getenv("K3S_AGENT_CPU", "0.1"), 64)
	config.K3SAgentMEM, _ = strconv.ParseFloat(util.Getenv("K3S_AGENT_MEM", "2000"), 64)
	config.K3SToken = util.Getenv("K3S_TOKEN", "123456789")
	config.LogLevel = util.Getenv("LOGLEVEL", "info")
	config.Principal = util.Getenv("MESOS_PRINCIPAL", "")
	config.ImageK3S = util.Getenv("IMAGE_K3S", "avhost/ubuntu-m3s:focal")
	config.ImageETCD = util.Getenv("IMAGE_ETCD", "docker.io/bitnami/etcd:3.5.1")
	config.ImageMySQL = util.Getenv("IMAGE_MYSQL", "docker.io/mariadb:10.8.3")
	config.RedisServer = util.Getenv("REDIS_SERVER", "127.0.0.1:6379")
	config.RedisPassword = util.Getenv("REDIS_PASSWORD", "")
	config.RedisDB, _ = strconv.Atoi(util.Getenv("REDIS_DB", "1"))
	config.SSLKey = util.Getenv("SSL_KEY_BASE64", "")
	config.SSLCrt = util.Getenv("SSL_CRT_BASE64", "")
	config.SkipSSL = stringToBool(util.Getenv("SKIP_SSL", "true"))
	config.VolumeDriver = util.Getenv("VOLUME_DRIVER", "local")
	config.VolumeK3SServer = util.Getenv("VOLUME_K3S_SERVER", "/data/k3s/server")
	config.VolumeDS = util.Getenv("VOLUME_DS", "/data/k3s/datastore")

	// if labels are set, unmarshel it into the Mesos Label format.
	labels := os.Getenv("K3S_AGENT_LABELS")
	if labels != "" {
		err := json.Unmarshal([]byte(labels), &config.K3SAgentLabels)

		if err != nil {
			logrus.Fatal("The env variable K3S_AGENT_LABELS have a syntax failure: ", err)
		}
	}

	// if the constraint is set, determine which kind of
	config.K3SServerConstraint = util.Getenv("K3S_SERVER_CONSTRAINT", "")
	if strings.Contains(config.K3SServerConstraint, ":") {
		constraint := strings.Split(config.K3SServerConstraint, ":")

		switch strings.ToLower(constraint[0]) {
		case "hostname":
			config.K3SServerConstraintHostname = strings.ToLower(constraint[1])
		}
	}

	// if the constraint is set, determine which kind of
	config.K3SAgentConstraint = util.Getenv("K3S_AGENT_CONSTRAINT", "")
	if strings.Contains(config.K3SAgentConstraint, ":") {
		constraint := strings.Split(config.K3SAgentConstraint, ":")

		switch strings.ToLower(constraint[0]) {
		case "hostname":
			config.K3SAgentConstraintHostname = strings.ToLower(constraint[1])
		}
	}

	// if the constraint is set, determine which kind of
	config.DSConstraint = util.Getenv("K3S_DS_CONSTRAINT", "")
	if strings.Contains(config.DSConstraint, ":") {
		constraint := strings.Split(config.DSConstraint, ":")

		switch strings.ToLower(constraint[0]) {
		case "hostname":
			config.DSConstraintHostname = strings.ToLower(constraint[1])
		}
	}

	// Enable Docker Engine vor K3S instead critc
	if strings.Compare(os.Getenv("K3S_DOCKER"), "true") == 0 {
		config.K3SDocker = " --container-runtime-endpoint unix:///var/run/cri-dockerd.sock "
	} else {
		config.K3SDocker = ""
	}

	// check if the domain starts with dot. if not, add one.
	if !strings.HasPrefix(config.Domain, ".") {
		tmp := config.Domain
		config.Domain = "." + tmp
	}
}

func stringToBool(par string) bool {
	return strings.Compare(par, "true") == 0
}
