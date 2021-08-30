package api

import (
	"net/http"

	mesos "github.com/AVENTER-UG/mesos-m3s/mesos"

	"github.com/sirupsen/logrus"
)

// V0ReflateK3SAgent will scale the agent service
// example:
// curl -X GET 127.0.0.1:10000/v0/agent/reflate -d 'JSON'
func V0ReflateK3SAgent(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("HTTP GET V0RestartMissingK3SAgent")
	auth := CheckAuth(r, w)

	if !auth {
		return
	}

	mesos.Revive()

	mesos.SearchMissingK3SAgent()

	//	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write([]byte("ok"))
}
