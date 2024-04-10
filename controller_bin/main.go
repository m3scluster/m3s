package main

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AVENTER-UG/mesos-m3s/controller/redis"
	cfg "github.com/AVENTER-UG/mesos-m3s/controller/types"
	framework "github.com/AVENTER-UG/mesos-m3s/types"
	util "github.com/AVENTER-UG/util/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	// BuildVersion of m3s
	BuildVersion string
	// GitVersion is the revision and commit number
	GitVersion string
	// VersionURL is the URL of the .version.json file
	VersionURL string
	// Config is the configuration of this application
	Config cfg.Config
	// Redis handler
	Redis *redis.Redis
	// Client the kubernetes client object
	Client *kubernetes.Clientset
	// Framework config
	Framework framework.Config
	// ReadyNodes is the amount of ready K8 nodes
	ReadyNodes int
)

// Heartbeat - call several function after the configure time
func Heartbeat() {
	ticker := time.NewTicker(Config.Heartbeat)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		logrus.WithField("func", "controller.Heartbeat").Debug("Heartbeat")
		// update framework config
		go loadFrameworkConfig()
	}
}

func loadFrameworkConfig() {
	Redis = redis.New(&Config)
	Redis.Connect()
	time.Sleep(60 * time.Second)

	key := Redis.GetRedisKey(Config.RedisPrefix + ":framework_config")
	if key != "" {
		json.Unmarshal([]byte(key), &Framework)
	}
}

func startController() {
	// load framework config
	loadFrameworkConfig()

	// get kubeconfig
	kubeconfig, err := config.GetConfig()
	if err != nil {
		logrus.WithField("func", "controller.startController").Error(err, "unable to set get kubeconfig")
		return
	}

	// get kubeconfig and store in in REDIS
	content, err := os.ReadFile(Config.KubernetesConfig)
	if err != nil {
		logrus.WithField("func", "controller.startController").Error("Error reading file:", err)
	} else {
		Redis.SetRedisKey(content, Config.RedisPrefix+":kubernetes_config")
		loadServerToken()
	}

	// Create kubernetes client
	Client, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		logrus.WithField("func", "controller.startController").Error(err, "unable to create kubernetes client")
		return
	}

	waitForKubernetesMasterReady(Client)
	loadDefaultYAML()
}

func loadDefaultYAML() {
	logrus.WithField("func", "loadDefaultYAML").Info("Load default yaml to apply: ", Config.DefaultYAML)

	yamlFile, err := os.ReadFile(Config.DefaultYAML)
	if err != nil {
		logrus.WithField("func", "controller.loadDefaultYAML").Error("could not load default YAML file: ", Config.DefaultYAML)
		return
	}

	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, _, err = decoder.Decode(yamlFile, nil, obj)
	if err != nil {
		logrus.WithField("func", "controller.loadDefaultYAML").Error("Error during encode default YAML file: ", err.Error())
		return
	}

	Client.RESTClient().
		Post().
		Resource(obj.GetKind()).
		Namespace(obj.GetNamespace()).
		Body(obj).Do(context.TODO())
}

func loadServerToken() {
	logrus.WithField("func", "loadServerToken").Info("Load K3s Server Token: ", Config.ServerTokenPath)

	tokenFile, err := os.ReadFile(Config.ServerTokenPath)
	if err != nil {
		logrus.WithField("func", "controller.loadServerToken").Error("could not load server token file: ", Config.ServerTokenPath)
		return
	}
	Redis.SetRedisKey(tokenFile, Config.RedisPrefix+":kubernetes_servertoken")
}

func waitForKubernetesMasterReady(clientset *kubernetes.Clientset) {
	logrus.WithField("func", "controller.waitForKubernetesMasterReady").Info("Wait until Kubernetes Manager is ready")
	for {
		_, err := clientset.ServerVersion()
		if err == nil {
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func init() {
	Config.AppName = "M3s Kubernetes Controller"
	Config.LogLevel = util.Getenv("M3S_CONTROLLER__LOGLEVEL", "DEBUG")
	Config.RedisServer = util.Getenv("M3S_CONTROLLER__REDIS_SERVER", "127.0.0.1:6480")
	Config.RedisPassword = util.Getenv("M3S_CONTROLLER__REDIS_PASSWORD", "")
	Config.RedisDB, _ = strconv.Atoi(util.Getenv("M3S_CONTROLLER__REDIS_DB", "1"))
	Config.RedisPrefix = util.Getenv("M3S_CONTROLLER__REDIS_PREFIX", "m3s")
	Config.KubernetesConfig = util.Getenv("KUBECONFIG", "/etc/rancher/k3s/k3s.yaml")
	Config.DefaultYAML = util.Getenv("M3S_CONTROLLER__DEFAULT_YAML", "/mnt/mesos/sandbox/default.yaml")
	Config.Heartbeat, _ = time.ParseDuration(util.Getenv("M3S_CONTROLLER__HEARTBEAT_TIME", "2m"))
	Config.UnscheduleTime, _ = time.ParseDuration(util.Getenv("M3S_CONTROLLER__UNSCHEDULE_TIME", "10s"))
	Config.ServerTokenPath = util.Getenv("M3S_CONTROLLER__SERVER_TOKEN_PATH", "/var/lib/rancher/k3s/server/token")

	if strings.Compare(util.Getenv("M3S_CONTROLLER__ENABLE_TAINT", "true"), "false") == 0 {
		Config.EnableTaint = false
	} else {
		Config.EnableTaint = true
	}

	util.SetLogging(Config.LogLevel, false, Config.AppName)
	logrus.Println(Config.AppName + " build " + BuildVersion + " git " + GitVersion)
}

func main() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		startController()
		content := Redis.GetRedisKey(Config.RedisPrefix + ":kubernetes_config")
		if content != "" {
			ticker.Stop()
		}
	}

}
