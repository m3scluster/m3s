package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AVENTER-UG/mesos-m3s/controller/redis"
	cfg "github.com/AVENTER-UG/mesos-m3s/controller/types"
	util "github.com/AVENTER-UG/util/util"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
)

// reconcileReplicaSet reconciles ReplicaSets
type reconcileReplicaSet struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
}

// Heartbeat - call several function after the configure time
func Heartbeat() {
	ticker := time.NewTicker(Config.Heartbeat)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		logrus.WithField("func", "controller.Heartbeat").Debug("Heartbeat")
		go CleanupNodes()
	}
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &reconcileReplicaSet{}

// Reconcile after kubernetes events happen. This function will store node information in redis.
func (r *reconcileReplicaSet) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	node := &corev1.Node{}
	err := r.client.Get(ctx, request.NamespacedName, node)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not find pods: %s", err)
	}

	if Config.EnableTaint {
		r.setTaint(node)
	}

	nodeName := node.ObjectMeta.Name
	logrus.WithField("func", "controller.Reconciler").Debug(nodeName)
	d, _ := json.Marshal(&node)
	Redis.SetRedisKey(d, Config.RedisPrefix+":kubernetes:"+nodeName)

	err = r.client.Update(ctx, node)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not write node: %s", err)
	}

	return reconcile.Result{}, nil
}

// setTain will prevent to run other pods then the control-plane on the kubernetes master server
func (r *reconcileReplicaSet) setTaint(node *corev1.Node) {
	for i := range node.Labels {
		if i == "node-role.kubernetes.io/master" {
			taint := corev1.Taint{
				Key:    "node-role.kubernetes.io/master",
				Value:  "NoSchedule",
				Effect: corev1.TaintEffectNoSchedule,
			}

			if !r.taintExist(node.Spec.Taints, "node-role.kubernetes.io/master", "NoSchedule", corev1.TaintEffectNoSchedule) {
				logrus.WithField("func", "controller.setTaint").Debug("Set Taint on: ", node.ObjectMeta.Name)
				node.Spec.Taints = append(node.Spec.Taints, taint)
			}
		}
	}
}

// taintExist will check (true) if the given taint already exist
func (r *reconcileReplicaSet) taintExist(taint []corev1.Taint, key string, value string, effect corev1.TaintEffect) bool {
	for _, i := range taint {
		if i.Key == key && i.Value == value && i.Effect == effect {
			return true
		}
	}

	return false
}

// CleanupNodes will cleanup unready nodes
func CleanupNodes() {
	nodes, err := Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.WithField("func", "controller.CleanupNodes").Error(err.Error)
		return
	}

	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status != corev1.ConditionTrue {
				// remove unready nodes from kubernetes and redis db
				logrus.WithField("func", "controller.CleanupNodes").Info("Remove unready node: ", node.Name)
				Client.CoreV1().Nodes().Delete(context.Background(), node.Name, metav1.DeleteOptions{})
			}
		}
	}
}

func startController() {
	Redis = redis.New(&Config)
	Redis.Connect()
	time.Sleep(60 * time.Second)

	// get kubeconfig
	kubeconfig, err := config.GetConfig()
	if err != nil {
		logrus.WithField("func", "controller.startController").Error(err, "unable to set get kubeconfig")
		return
	}

	// Setup a Manager
	logrus.WithField("func", "controller.startController").Info("setting up manager")
	mgr, err := manager.New(kubeconfig, manager.Options{})
	if err != nil {
		logrus.WithField("func", "controller.startController").Error(err, "unable to set up overall controller manager")
		return
	}

	// get kubeconfig and store in in REDIS
	content, err := os.ReadFile(Config.KubernetesConfig)
	if err != nil {
		logrus.WithField("func", "controller.startController").Error("Error reading file:", err)
	} else {
		Redis.SetRedisKey(content, Config.RedisPrefix+":kubernetes_config")
	}

	// Create kubernetes client
	Client, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		logrus.WithField("func", "controller.startController").Error(err, "unable to create kubernetes node")
		return
	}

	waitForKubernetesMasterReady(Client)
	logrus.WithField("func", "controller.startController").Info("Kubernetes Manager is ready")

	// Setup a new controller to reconcile ReplicaSets
	logrus.WithField("func", "controller.startController").Info("Setting up controller")
	c, err := controller.New("m3s_controller", mgr, controller.Options{
		Reconciler: &reconcileReplicaSet{client: mgr.GetClient()},
	})

	// Watch for Node events and call Reconcile
	err = c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		logrus.WithField("func", "controller.startController").Error(err, "unable to watch Node")
		return
	}

	go Heartbeat()
	go loadDefaultYAML()

	logrus.WithField("func", "main").Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logrus.WithField("func", "controller.startController").Error(err, "unable to run manager")
		return
	}
}

func loadDefaultYAML() {
	logrus.WithField("func", "loadDefaultYAML").Info("Load default yaml to apply: ", Config.DefaultYAML)

	yamlFile, err := ioutil.ReadFile(Config.DefaultYAML)
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

	if strings.Compare(util.Getenv("M3S_CONTROLLER__ENABLE_TAINT", "true"), "false") == 0 {
		Config.EnableTaint = false
	} else {
		Config.EnableTaint = true
	}

	util.SetLogging(Config.LogLevel, false, Config.AppName)
	logrus.Println(Config.AppName + " build " + BuildVersion + " git " + GitVersion)
}

func main() {
	//	this loop is for reconnect purpose
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	// nolint:gosimple
	for {
		select {
		case <-ticker.C:
			startController()
		}
	}
}
