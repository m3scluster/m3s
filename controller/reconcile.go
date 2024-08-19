package controller

import (
	"context"
	"encoding/json"
	"fmt"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileReplicaSet reconciles ReplicaSets
type reconcileReplicaSet struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
	// ctx Kubernetes Context
	ctx context.Context

	control *Controller
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &reconcileReplicaSet{}

// Reconcile after kubernetes events happen. This function will store node information in redis.
func (r *reconcileReplicaSet) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.ctx = ctx
	node := &corev1.Node{}
	err := r.client.Get(ctx, request.NamespacedName, node)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not find pods: %s", err)
	}

	if r.control.Config.K3SEnableTaint {
		r.setTaint(node)
	}

	nodeName := node.ObjectMeta.Name
	logrus.WithField("func", "controller.Reconciler").Debug(nodeName)

	d, _ := json.Marshal(&node)
	r.control.Redis.SetRedisKey(d, r.control.Redis.Prefix+":kubernetes:"+node.ObjectMeta.Name)

	return reconcile.Result{}, nil
}

// update node will update the node in K8
func (r *reconcileReplicaSet) updateNode(node *corev1.Node) {
	err := r.client.Update(r.ctx, node)
	if err != nil {
		logrus.WithField("func", "controller.updateNode").Errorf("could not write node: %s", err)
		return
	}
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
				r.updateNode(node)
			} else {
				logrus.WithField("func", "controller.setTaint").Debug("Taint Already Exists on: ", node.ObjectMeta.Name)
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
