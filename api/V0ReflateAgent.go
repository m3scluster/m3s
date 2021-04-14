package api

import (
	"net/http"

	mesos "mesos-k3s/mesos"
	mesosproto "mesos-k3s/proto"

	"github.com/sirupsen/logrus"
)

// V0ReflateMissingK3SAgent will scale the agent service
// example:
// curl -X GET 127.0.0.1:10000/v0/agent/reflate -d 'JSON'
func V0ReflateK3SAgent(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("HTTP GET V0RestartMissingK3SAgent")
	auth := CheckAuth(r, w)

	if !auth {
		return
	}

	revive := &mesosproto.Call{
		Type: mesosproto.Call_REVIVE,
	}
	mesos.Call(revive)

	mesos.SearchMissingK3SAgent()

	//	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write([]byte("ok"))
}
