package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"plugin"
	"strconv"
	"strings"
	"time"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	util "github.com/AVENTER-UG/util/util"
	"github.com/Showmax/go-fqdn"

	"github.com/AVENTER-UG/mesos-m3s/redis"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
)

var config cfg.Config
var framework cfg.FrameworkConfig

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
	config.AppName = "Apache Mesos K3S Framework"
	config.BootstrapURL = util.Getenv("BOOTSTRAP_URL", "https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/master/bootstrap/bootstrap.sh")
	config.CGroupV2 = stringToBool(util.Getenv("CGROUP_V2", "false"))
	config.DSMax, _ = strconv.Atoi(util.Getenv("DS_COUNT", "1"))
	config.EventLoopTime, _ = time.ParseDuration(util.Getenv("HEARTBEAT_INTERVAL", "15s"))
	config.CleanupLoopTime, _ = time.ParseDuration(util.Getenv("CLEANUP_WAIT", "5m"))
	config.Credentials.Username = util.Getenv("AUTH_USERNAME", "")
	config.Credentials.Password = util.Getenv("AUTH_PASSWORD", "")
	config.Domain = util.Getenv("DOMAIN", ".local")
	config.DockerSock = os.Getenv("DOCKER_SOCK")
	config.DockerSHMSize = util.Getenv("DOCKER_SHM_SIZE", "30gb")
	config.DockerUlimit = util.Getenv("DOCKER_ULIMIT", "1048576")
	config.DockerMemorySwap, _ = strconv.ParseFloat(util.Getenv("DOCKER_MEMORY_SWAP", "1000"), 64)
	config.DockerCNI = util.Getenv("DOCKER_CNI", "bridge")
	config.DockerRunning = strings.Compare(os.Getenv("DOCKER_RUNNING"), "true") == 0
	config.DSCPU, _ = strconv.ParseFloat(util.Getenv("DS_CPU", "0.1"), 64)
	config.DSMEM, _ = strconv.ParseFloat(util.Getenv("DS_MEM", "1000"), 64)
	config.DSDISK, _ = strconv.ParseFloat(util.Getenv("DS_DISK", "10000"), 64)
	config.DSEtcd = stringToBool(util.Getenv("DS_ETCD", "false"))
	config.DSPort = util.Getenv("DS_PORT", "3306")
	config.DSMySQL = stringToBool(util.Getenv("DS_MYSQL", "true"))
	config.DSMySQLUsername = util.Getenv("DS_MYSQL_USERNAME", "root")
	config.DSMySQLPassword = util.Getenv("DS_MYSQL_PASSWORD", "password")
	config.DSMySQLSSL = stringToBool(util.Getenv("DS_MYSQL_SSL", "false"))
	config.KubeConfig = util.Getenv("KUBECONFIG", "/etc/rancher/k3s/k3s.yaml")
	config.KubernetesVersion = util.Getenv("KUBERNETES_VERSION", "v1.25.2")
	config.K3SNodeTimeout, _ = time.ParseDuration(util.Getenv("K3S_NODE_TIMEOUT", "10m"))
	config.K3SServerMax, _ = strconv.Atoi(util.Getenv("K3S_SERVER_COUNT", "1"))
	config.K3SServerPort, _ = strconv.Atoi(util.Getenv("K3S_SERVER_PORT", strconv.Itoa(framework.PortRangeFrom)))
	config.K3SServerCPU, _ = strconv.ParseFloat(util.Getenv("K3S_SERVER_CPU", "1.0"), 64)
	config.K3SServerMEM, _ = strconv.ParseFloat(util.Getenv("K3S_SERVER_MEM", "2000"), 64)
	config.K3SServerDISK, _ = strconv.ParseFloat(util.Getenv("K3S_SERVER_DISK", "1000"), 64)
	config.K3SServerString = util.Getenv("K3S_SERVER_STRING", "/usr/local/bin/k3s server --cgroup-driver=cgroupfs --cluster-cidr=10.2.0.0/16 --service-cidr=10.3.0.0/16 --cluster-dns=10.3.0.10 --kube-scheduler-arg=leader-elect=false --kube-controller-manager-arg=enable-leader-migration=false --kube-cloud-controller-manager-arg=enable-leader-migration=false --kube-controller-manager-arg=leader-elect=false --disable-cloud-controller --snapshotter=native --flannel-backend=vxlan")
	config.K3SCustomDomain = util.Getenv("K3S_CUSTOM_DOMAIN", "cloud.local")
	config.K3SContainerDisk = util.Getenv("K3S_CONTAINER_DISK", config.DockerSHMSize)
	config.K3SAgentString = util.Getenv("K3S_AGENT_STRING", "/usr/local/bin/k3s agent --snapshotter=native --cgroup-driver=cgroupfs")
	config.K3SAgentMax, _ = strconv.Atoi(util.Getenv("K3S_AGENT_COUNT", "1"))
	config.K3SAgentCPU, _ = strconv.ParseFloat(util.Getenv("K3S_AGENT_CPU", "2.0"), 64)
	config.K3SAgentMEM, _ = strconv.ParseFloat(util.Getenv("K3S_AGENT_MEM", "2000"), 64)
	config.K3SAgentDISK, _ = strconv.ParseFloat(util.Getenv("K3S_AGENT_DISK", "10000"), 64)
	config.K3SAgentTCPPort, _ = strconv.Atoi(util.Getenv("K3S_AGENT_TCP_PORT", "0"))
	config.K3SToken = util.Getenv("K3S_TOKEN", "123456789")
	config.LogLevel = util.Getenv("LOGLEVEL", "DEBUG")
	config.Principal = util.Getenv("MESOS_PRINCIPAL", "")
	config.ImageK3S = util.Getenv("IMAGE_K3S", "avhost/ubuntu-m3s:22.04-3")
	config.ImageETCD = util.Getenv("IMAGE_ETCD", "quay.io/coreos/etcd:v3.5.1")
	config.ImageMySQL = util.Getenv("IMAGE_MYSQL", "docker.io/mariadb:10.8.3")
	config.ReconcileLoopTime, _ = time.ParseDuration(util.Getenv("RECONCILE_WAIT", "10m"))
	config.RefuseOffers, _ = strconv.ParseFloat(util.Getenv("REFUSE_OFFERS", "60.0"), 64)
	config.ReviveLoopTime, _ = time.ParseDuration(util.Getenv("REVIVE_WAIT", "1m"))
	config.RedisServer = util.Getenv("REDIS_SERVER", "127.0.0.1:6379")
	config.RedisPassword = util.Getenv("REDIS_PASSWORD", "")
	config.RedisDB, _ = strconv.Atoi(util.Getenv("REDIS_DB", "1"))
	config.SSLKey = util.Getenv("SSL_KEY_BASE64", "")
	config.SSLCrt = util.Getenv("SSL_CRT_BASE64", "")
	config.SkipSSL = stringToBool(util.Getenv("SKIP_SSL", "true"))
	config.VolumeDriver = util.Getenv("VOLUME_DRIVER", "local")
	config.VolumeK3SServer = util.Getenv("VOLUME_K3S_SERVER", "/data/k3s/server")
	config.VolumeDS = util.Getenv("VOLUME_DS", "/data/k3s/datastore")
	config.TimeZone = util.Getenv("TZ", "Etc/UTC")
	config.K3SDisableScheduling = stringToBool(util.Getenv("K3S_DISABLE_SCHEDULING", "false"))
	config.DSMaxRestore = 0
	config.K3SAgentMaxRestore = 0
	config.K3SServerMaxRestore = 0

	// if agent labels are set, unmarshel it into the Mesos Label format.
	labels := os.Getenv("K3S_AGENT_LABELS")
	if labels != "" {
		err := json.Unmarshal([]byte(labels), &config.K3SAgentLabels)

		if err != nil {
			logrus.WithField("func", "init").Fatal("The env variable K3S_AGENT_LABELS have a syntax failure: ", err)
		}
	}

	// if server labels are set, unmarshel it into the Mesos Label format.
	labels = os.Getenv("K3S_SERVER_LABELS")
	if labels != "" {
		err := json.Unmarshal([]byte(labels), &config.K3SServerLabels)

		if err != nil {
			logrus.WithField("func", "init").Fatal("The env variable K3S_SERVER_LABELS have a syntax failure: ", err)
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
		config.K3SDocker = "--container-runtime-endpoint unix:///var/run/cri-dockerd.sock"
	} else {
		config.K3SDocker = ""
	}

	// check if the domain starts with dot. if not, add one.
	if !strings.HasPrefix(config.Domain, ".") {
		tmp := config.Domain
		config.Domain = "." + tmp
	}

	if strings.Compare(util.Getenv("K3S_ENABLE_TAINT", "true"), "false") == 0 {
		config.K3SEnableTaint = false
	} else {
		config.K3SEnableTaint = true
	}

	// Enable plugins
	if strings.Compare(util.Getenv("M3S_PLUGINS", "false"), "false") == 0 {
		config.PluginsEnable = false
	} else {
		config.PluginsEnable = true
	}
}

func loadPlugins(r *redis.Redis) {
	if config.PluginsEnable {
		config.Plugins = map[string]*plugin.Plugin{}

		plugins, err := filepath.Glob("plugins/*.so")
		if err != nil {
			logrus.WithField("func", "main.loadPlugins").Info("No Plugins found")
			return
		}

		for _, filename := range plugins {
			p, err := plugin.Open(filename)
			if err != nil {
				logrus.WithField("func", "main.initPlugins").Error("Error during loading plugin: ", err.Error())
				continue
			}

			symbol, err := p.Lookup("Init")
			if err != nil {
				logrus.WithField("func", "main.initPlugins").Error("Error lookup init plugin: ", err.Error())
				continue
			}

			initPluginFunc, ok := symbol.(func(*redis.Redis) string)

			if !ok {
				logrus.WithField("func", "main.initPlugins").Error("Error plugin does not have init function")
				continue
			}

			name := initPluginFunc(r)
			config.Plugins[name] = p
		}
		logrus.SetPlugins(config.Plugins)
	}
}

func stringToBool(par string) bool {
	return strings.Compare(par, "true") == 0
}
