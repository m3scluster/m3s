package api

import (
	"encoding/json"
	"net/http"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
)

// V0GetCountK3SServer will write out the current count of Kubernetes server and the expected one.
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/server/scale'
func (e *API) V0GetCountK3SServer(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0GetCountK3SServer").Debug("Count K3S Server")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var count cfg.Count

	count.Scale = e.Config.K3SServerMax
	count.Running = e.Redis.CountRedisKey(e.Framework.FrameworkName+":server:*", "")

	d, err := json.Marshal(count)

	if err != nil {
		logrus.WithField("func", "api.V0GetCountK3SServer").Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
