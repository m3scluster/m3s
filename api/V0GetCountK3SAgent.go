package api

import (
	"encoding/json"
	"net/http"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
)

// V0GetCountK3SAgent will write out the current count of Kubernetes Agents and the expected one.
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/agent/scale'
func (e *API) V0GetCountK3SAgent(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0GetCountK3SAgent").Debug("Count K3S Agents")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var count cfg.Count

	count.Scale = e.Config.K3SAgentMax
	count.Running = e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", "")

	d, err := json.Marshal(count)

	if err != nil {
		logrus.WithField("func", "api.V0GetCountK3SAgent").Error(err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
