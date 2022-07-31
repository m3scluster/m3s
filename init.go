package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	util "github.com/AVENTER-UG/util"
	"github.com/Showmax/go-fqdn"
	"github.com/sirupsen/logrus"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"
)

var config cfg.Config
var framework mesosutil.FrameworkConfig

func init() {
	config.K3SAgentMax = 0
	config.K3SServerMax = 0

	framework.FrameworkUser = util.Getenv("FRAMEWORK_USER", "root")
	framework.FrameworkName = "m3s" + util.Getenv("FRAMEWORK_NAME", "")
	framework.FrameworkRole = util.Getenv("FRAMEWORK_ROLE", "m3s")
	framework.FrameworkPort = util.Getenv("FRAMEWORK_PORT", "10000")
	framework.FrameworkHostname = util.Getenv("FRAMEWORK_HOSTNAME", fqdn.Get())
	framework.FrameworkInfoFilePath = util.Getenv("FRAMEWORK_STATEFILE_PATH", "/tmp")
	framework.Username = os.Getenv("MESOS_USERNAME")
	framework.Password = os.Getenv("MESOS_PASSWORD")
	framework.MesosMasterServer = util.Getenv("MESOS_MASTER", "127.0.0.1:5050")
	framework.MesosCNI = os.Getenv("MESOS_CNI")
	framework.PortRangeFrom, _ = strconv.Atoi(util.Getenv("PORTRANGE_FROM", "31000"))
	framework.PortRangeTo, _ = strconv.Atoi(util.Getenv("PORTRANGE_TO", "32000"))
	config.Principal = os.Getenv("MESOS_PRINCIPAL")
	config.LogLevel = util.Getenv("LOGLEVEL", "info")
	config.Domain = util.Getenv("DOMAIN", ".local")
	config.K3SAgentMax, _ = strconv.Atoi(util.Getenv("K3S_AGENT_COUNT", "1"))
	config.K3SServerMax, _ = strconv.Atoi(util.Getenv("K3S_SERVER_COUNT", "1"))
	config.Credentials.Username = os.Getenv("AUTH_USERNAME")
	config.Credentials.Password = os.Getenv("AUTH_PASSWORD")
	config.AppName = "Mesos K3S Framework"
	config.K3SCustomDomain = util.Getenv("K3S_CUSTOM_DOMAIN", "cloud.local")
	config.K3SServerString = util.Getenv("K3S_SERVER_STRING", "/usr/local/bin/k3s server --cluster-cidr=10.2.0.0/16 --service-cidr=10.3.0.0/16 --cluster-dns=10.3.0.10  --kube-controller-manager-arg='leader-elect=false' --disable-cloud-controller --kube-scheduler-arg='leader-elect=false' --snapshotter=native --flannel-backend=vxlan ")
	config.K3SAgentString = util.Getenv("K3S_AGENT_STRING", "/usr/local/bin/k3s agent --snapshotter=native --flannel-backend=vxlan ")
	config.ImageK3S = util.Getenv("IMAGE_K3S", "avhost/ubuntu-m3s:focal")
	config.ImageETCD = util.Getenv("IMAGE_ETCD", "docker.io/bitnami/etcd:3.5.1")
	config.ImageMySQL = util.Getenv("IMAGE_MYSQL", "docker.io/mariadb:10.8.3")
	config.VolumeDriver = util.Getenv("VOLUME_DRIVER", "local")
	config.VolumeK3SServer = util.Getenv("VOLUME_K3S_SERVER", "/data/k3s")
	config.K3SToken = util.Getenv("K3S_TOKEN", "123456789")
	config.DSMax, _ = strconv.Atoi(util.Getenv("DS_COUNT", "1"))
	config.BootstrapURL = util.Getenv("BOOTSTRAP_URL", "https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/dev/bootstrap/bootstrap.sh")
	config.DockerSock = os.Getenv("DOCKER_SOCK")
	config.DockerSHMSize = util.Getenv("DOCKER_SHM_SIZE", "30gb")
	config.M3SBootstrapServerPort, _ = strconv.Atoi(util.Getenv("M3S_BOOTSTRAP_SERVER_PORT", "6443"))
	config.K3SServerCPU, _ = strconv.ParseFloat(util.Getenv("K3S_SERVER_CPU", "0.1"), 64)
	config.K3SServerMEM, _ = strconv.ParseFloat(util.Getenv("K3S_SERVER_MEM", "1200"), 64)
	config.K3SAgentCPU, _ = strconv.ParseFloat(util.Getenv("K3S_AGENT_CPU", "0.1"), 64)
	config.K3SAgentMEM, _ = strconv.ParseFloat(util.Getenv("K3S_AGENT_MEM", "1200"), 64)
	config.DSCPU, _ = strconv.ParseFloat(util.Getenv("DS_CPU", "0.1"), 64)
	config.DSMEM, _ = strconv.ParseFloat(util.Getenv("DS_MEM", "100"), 64)
	config.DSDISK, _ = strconv.ParseFloat(util.Getenv("DS_DISK", "10000"), 64)
	config.DSPort = util.Getenv("DS_PORT", "3306")
	config.DSMySQLUsername = util.Getenv("DS_MYSQL_USERNAME", "root")
	config.DSMySQLPassword = util.Getenv("DS_MYSQL_PASSWORD", "password")
	config.RedisServer = util.Getenv("REDIS_SERVER", "127.0.0.1:6379")
	config.RedisPassword = os.Getenv("REDIS_PASSWORD")
	config.RedisDB, _ = strconv.Atoi(util.Getenv("REDIS_DB", "1"))
	config.SSLKey = os.Getenv("SSL_KEY_BASE64")
	config.SSLCrt = os.Getenv("SSL_CRT_BASE64")
	config.DockerCNI = util.Getenv("DOCKER_CNI", "bridge")

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
		config.K3SDocker = " --docker "
	} else {
		config.K3SDocker = ""
	}

	// The comunication to the mesos server should be via ssl or not
	framework.MesosSSL = stringToBool(os.Getenv("MESOS_SSL"))

	// Skip SSL Verification
	config.SkipSSL = stringToBool(os.Getenv("SKIP_SSL"))

	// Set the kind of datastore endpoint
	config.DSEtcd = stringToBool(os.Getenv("DS_ETCD"))
	config.DSMySQL = stringToBool(util.Getenv("DS_MYSQL", "true"))

	// check if the domain starts with dot. if not, add one.
	if !strings.HasPrefix(config.Domain, ".") {
		tmp := config.Domain
		config.Domain = "." + tmp
	}
}

func stringToBool(par string) bool {
	if strings.Compare(par, "true") == 0 {
		return true
	}
	return false
}
