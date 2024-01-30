package api

import (
	"net/http"
	"strconv"
	"strings"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// V0GetKubeconfig will return the kubernetes config file
// example:
// curl -X GET 127.0.0.1:10000/api/m3s/v0/server/config"
func (e *API) V0GetKubeconfig(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0GetKubeconfig").Debug("Call")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	content := e.Redis.GetRedisKey(e.Framework.FrameworkName + ":kubernetes_config")

	if content != "" {
		// replace the localhost server string with the mesos agent hostname and dynamic port
		destURL := e.Config.K3SServerHostname + ":" + strconv.Itoa(e.Config.K3SServerPort)
		kubconf := strings.Replace(string(content), "127.0.0.1:6443", destURL, -1)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Api-Service", "v0")
		w.Write([]byte(kubconf))
	}
}
