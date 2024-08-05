package controller

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/AVENTER-UG/mesos-m3s/mesos"
	"github.com/AVENTER-UG/mesos-m3s/redis"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Controller struct {
	Config     *cfg.Config
	Framework  *cfg.FrameworkConfig
	Redis      *redis.Redis
	Client     *kubernetes.Clientset
	Kubeconfig *rest.Config
	Mesos      mesos.Mesos
	ReadyNodes int
}

// New will create a new Controller object
func New(cfg *cfg.Config, frm *cfg.FrameworkConfig) *Controller {
	e := &Controller{
		Config:    cfg,
		Framework: frm,
		Mesos:     *mesos.New(cfg, frm),
	}

	log.SetLogger(zap.New(zap.UseDevMode(false), zap.Level(zapcore.Level(4))))
	logrus.WithField("func", "Controller.New").Info("Create M3S-K8 Controller")

	go e.heartbeat()

	return e
}

func (e *Controller) CreateClient() {
	config := e.Redis.GetRedisKey(e.Redis.Prefix + ":kubernetes_config")

	for config == "" {
		logrus.WithField("func", "controller.CreateClient").Info("kubernetes_config does not exist. Retry!")
		config = e.Redis.GetRedisKey(e.Redis.Prefix + ":kubernetes_config")
		time.Sleep(30 * time.Second)
	}

	destURL := e.Config.K3SServerHostname + ":" + strconv.Itoa(e.Config.K3SServerPort)
	config = strings.Replace(string(config), "127.0.0.1:6443", destURL, -1)

	var err error
	e.Kubeconfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(config))
	if err != nil {
		logrus.WithField("func", "controller.CreateClient").Error("Load K8 config error: ", err.Error())
		return
	}

	e.Client, err = kubernetes.NewForConfig(e.Kubeconfig)
	if err != nil {
		logrus.WithField("func", "controller.CreateClient").Error("Create K8 clientset error: ", err.Error())
		return
	}
}

func (e *Controller) waitForKubernetesMasterReady() {
	logrus.WithField("func", "controller.waitForKubernetesMasterReady").Info("Wait until Kubernetes is ready")
	for {
		_, err := e.Client.ServerVersion()
		if err == nil {
			logrus.WithField("func", "controller.isReady").Info("Kubernetes is ready")
			return
		}
		time.Sleep(30 * time.Second)
	}
}

func (e *Controller) CreateController() {
	// Setup a Manager
	logrus.WithField("func", "controller.CreateController").Info("Create Controller")
	mgr, err := manager.New(e.Kubeconfig, manager.Options{})
	if err != nil {
		logrus.WithField("func", "controller.CreateController").Error(err, "unable to set up overall controller manager")
		return
	}

	logrus.WithField("func", "controller.CreateController").Info("Controller is ready")

	// Setup a new controller to reconcile ReplicaSets
	c, err := controller.New("m3s_controller", mgr, controller.Options{
		Reconciler: &reconcileReplicaSet{
			client:  mgr.GetClient(),
			control: e,
		},
	})
	if err != nil {
		logrus.WithField("func", "controller.startController").Error(err, "unable to create controller")
		return
	}

	// Watch for Node events and call Reconcile
	//err = c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{})

	err = c.Watch(source.Kind(mgr.GetCache(), &corev1.Node{}), &handler.EnqueueRequestForObject{})
	if err != nil {
		logrus.WithField("func", "controller.startController").Error(err, "unable to watch Node")
		return
	}

	logrus.WithField("func", "main.CreateController").Info("Starting Controller")
	if err := mgr.Start(context.Background()); err != nil {
		logrus.WithField("func", "controller.CreateController").Error(err, "unable to run controller")
		return
	}
}

// CleanupNodes will cleanup unready nodes
func (e *Controller) CleanupNodes() {
	if e.Client != nil {
		nodes, err := e.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			logrus.WithField("func", "controller.CleanupNodes").Error(err.Error())
			return
		}

		for _, node := range nodes.Items {
			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status != corev1.ConditionTrue {
					// remove unready nodes from kubernetes and redis db
					logrus.WithField("func", "controller.CleanupNodes").Info("Remove unready node: ", node.Name)
					e.DeleteNode(node.Name)
					task := e.GetTaskFromK8Node(node, "agent")
					if task.TaskID == "" {
						task = e.GetTaskFromK8Node(node, "server")
					}
					e.Redis.DelRedisKey(e.Framework.FrameworkName + ":kubernetes:" + node.Name)
					e.Mesos.Kill(task.TaskID, task.Agent)
				}
			}
		}
	}
}

func (e *Controller) updateServerLabel() {
	if e.Client != nil {
		nodes, err := e.Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			logrus.WithField("func", "controller.UpdateServerLabel").Error(err.Error())
			return
		}

		for _, node := range nodes.Items {
			if !strings.Contains(node.Name, "server") {
				continue
			}

			task := e.Redis.GetTaskByHostname(node.Name)
			if task.TaskID == "" {
				continue
			}

			labelKey := "m3s.aventer.biz/taskid"
			labelValue := task.TaskID

			if !e.isLabelEqual(node.Labels, labelKey, labelValue) {
				node.Labels[labelKey] = labelValue
				logrus.WithField("func", "controller.updateServerLabel").Infof("Update label (m3s.aventer.biz/taskid=%s) of %s", task.TaskID, node.Name)
			}

			_, err = e.Client.CoreV1().Nodes().Update(context.Background(), &node, metav1.UpdateOptions{})
			if err != nil {
				logrus.WithField("func", "controller.updateServerLabel").Info("Could not update node: ", err.Error())
				continue
			}
		}
	}
}

func (e *Controller) isLabelEqual(labels map[string]string, key string, value string) bool {
	for i, label := range labels {
		if i == key && label == value {
			logrus.WithField("func", "controller.GetTaskIDFromLabel").Tracef("Label %s TaskID %s", i, label)
			return true
		}
	}
	return false
}

func (e *Controller) DeleteNode(nodeName string) {
	e.Client.CoreV1().Nodes().Delete(context.Background(), nodeName, metav1.DeleteOptions{})
	e.CleanupNodes()
}

// Unschedule Node
func (e *Controller) CheckReadyState() bool {
	logrus.WithField("func", "controller.CheckReadyState").Trace("Unschedule")

	keys := e.Redis.GetAllRedisKeys(e.Redis.Prefix + ":kubernetes:*agent*")
	nodeReady := 0
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		var node corev1.Node
		err := json.NewDecoder(strings.NewReader(key)).Decode(&node)
		if err != nil {
			logrus.WithField("func", "controller.CheckReadyState").Error("Could not decode kubernetes node: ", err.Error())
			continue
		}
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionFalse {
				logrus.WithField("func", "controller.CheckReadyState").Trace("Unready")
				return false
			}
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				nodeReady++
			}
		}
	}

	if nodeReady != e.Config.K3SAgentMax {
		logrus.WithField("func", "controller.CheckReadyState").Trace("Unready")
		return false
	}

	logrus.WithField("func", "controller.CheckReadyState").Trace("Ready")
	return true
}

// SetUnscheduled set all nodes to unscheduled.
func (e *Controller) SetUnschedule() {
	keys := e.Redis.GetAllRedisKeys(e.Redis.Prefix + ":kubernetes:*agent*")
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		var node *corev1.Node
		err := json.NewDecoder(strings.NewReader(key)).Decode(&node)
		if err != nil {
			logrus.WithField("func", "controller.SetUnschedule").Error("Could not decode kubernetes node: ", err.Error())
			continue
		}
		if strings.Contains(node.ObjectMeta.Name, "server") {
			return
		}

		if !node.Spec.Unschedulable {
			node.Spec.Unschedulable = true
			_, err = e.Client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
			if err != nil {
				logrus.WithField("func", "controller.SetUnschedule").Info("Could not update node: ", err.Error())
				continue
			}
			logrus.WithField("func", "controller.SetUnscheduled").Infof("Unschedule Node: %s", node.ObjectMeta.Name)
		}
	}
}

// SetScheduled set all nodes to unscheduled.
func (e *Controller) SetSchedule() {
	keys := e.Redis.GetAllRedisKeys(e.Redis.Prefix + ":kubernetes:*agent*")
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		var node *corev1.Node
		err := json.NewDecoder(strings.NewReader(key)).Decode(&node)
		if err != nil {
			logrus.WithField("func", "controller.SetSchedule").Error("Could not decode kubernetes node: ", err.Error())
			continue
		}
		if strings.Contains(node.ObjectMeta.Name, "server") {
			return
		}

		if node.Spec.Unschedulable {
			node.Spec.Unschedulable = false
			_, err = e.Client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
			if err != nil {
				logrus.WithField("func", "controller.SetSchedule").Info("Could not update node: ", err.Error())
				continue
			}
			logrus.WithField("func", "controller.SetScheduled").Infof("Schedule Node: %s", node.ObjectMeta.Name)
		}
	}
}

// GetTaskFromK8Node will give out the mesos task matched to the K8 node
func (e *Controller) GetTaskFromK8Node(node corev1.Node, kind string) cfg.Command {
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":" + kind + ":*")
	for keys.Next(e.Redis.CTX) {
		taskID := e.GetTaskIDFromLabel(node.Labels)
		if taskID == "" {
			taskID = e.GetTaskIDFromAnnotation(node.Annotations)
		}
		if taskID != "" {
			key := e.Redis.GetRedisKey(e.Framework.FrameworkName + ":" + kind + ":" + taskID)
			if key != "" {
				task := e.Mesos.DecodeTask(key)
				return task
			} else {
				logrus.WithField("func", "controller.GetTaskFromK8Node").Tracef("Could not found key of taskID %s", taskID)
			}
		}
	}
	return cfg.Command{}
}

// GetK8NodeFromTask will give out the K8 node from mesos task
func (e *Controller) GetK8NodeFromTask(task cfg.Command) corev1.Node {
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":kubernetes:*")

	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		var k8Node corev1.Node
		err := json.NewDecoder(strings.NewReader(key)).Decode(&k8Node)
		if err != nil {
			logrus.WithField("func", "scheduler.getK8NodeFromTask").Trace("Could not decode kubernetes node: ", err.Error())
			continue
		}
		taskID := e.GetTaskIDFromLabel(k8Node.Labels)
		if taskID == "" {
			taskID = e.GetTaskIDFromAnnotation(k8Node.Annotations)
		}
		if taskID == task.TaskID {
			return k8Node
		}
	}

	return corev1.Node{}
}

// GetTaskIDFromLabel will return the Mesos Task ID in the label string
func (e *Controller) GetTaskIDFromLabel(labels map[string]string) string {
	for i, label := range labels {
		if i == "m3s.aventer.biz/taskid" {
			logrus.WithField("func", "controller.GetTaskIDFromLabel").Tracef("Label %s TaskID %s", i, label)
			return label
		}
	}
	return ""
}

// GetTaskIDFromAnnotation will return the Mesos Task ID in the annotation string
func (e *Controller) GetTaskIDFromAnnotation(annotations map[string]string) string {
	for i, annotation := range annotations {
		if i == "k3s.io/node-args" {
			var args []string
			err := json.Unmarshal([]byte(annotation), &args)
			if err != nil {
				logrus.WithField("func", "scheduler.getTaskIDFromAnnotation").Error("Could not decode kubernetes node annotation: ", err.Error())
				continue
			}
			for _, arg := range args {
				if strings.Contains(arg, "taskid") {
					value := strings.Split(arg, "=")
					logrus.WithField("func", "controller.GetTaskIDFromAnnotation").Tracef("Annotation %s", arg)
					if len(value) == 2 {
						return value[1]
					}
				}
			}
		}
	}
	return ""
}

func (e *Controller) heartbeat() {
	ticker := time.NewTicker(e.Config.EventLoopTime)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		logrus.WithField("func", "controller.Heartbeat").Debug("Heartbeat")
		// create controller and client if it does not exist
		if e.Client == nil {
			e.CreateClient()
			e.waitForKubernetesMasterReady()
			go e.CreateController()
		} else {
			e.updateServerLabel()
			if !e.CheckReadyState() || e.Config.K3SDisableScheduling {
				e.SetUnschedule()
			} else {
				e.SetSchedule()
			}
		}
	}
}
