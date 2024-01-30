package api

import (
	"encoding/json"
	"net/http"
	"strings"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	corev1 "k8s.io/api/core/v1"
)

// V0GetKubeVersion will return the kubernetes Version
// example:
// curl -X GET 127.0.0.1:10000/v0/server/version'
func (e *API) V0GetKubeVersion(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0GetKubeVersion").Debug("Call")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":kubernetes:*")

	var k3sVersion []cfg.K3SVersion

	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		var node corev1.Node
		err := json.NewDecoder(strings.NewReader(key)).Decode(&node)
		if err != nil {
			logrus.WithField("func", "scheduler.V0GetKubeVersion").Error("Could not decode kubernetes node: ", err.Error())
			continue
		}

		tmpVersion := cfg.K3SVersion{}
		tmpVersion.NodeName = node.Name
		tmpVersion.NodeInfo = node.Status.NodeInfo

		k3sVersion = append(k3sVersion, tmpVersion)
	}

	e.Config.Version.K3SVersion = k3sVersion

	d, _ := json.Marshal(&e.Config.Version)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}
