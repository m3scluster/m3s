package api

import (
	"encoding/json"
	"net/http"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	"github.com/sirupsen/logrus"
)

// V0GetCountK3SServer will write out the current count of Kubernetes server and the expected one.
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/server/scale'
func (e *API) V0GetCountK3SServer(w http.ResponseWriter, r *http.Request) {
	auth := e.CheckAuth(r, w)

	if !auth {
		return
	}

	var count cfg.Count

	count.Scale = e.Config.K3SServerMax
	count.Running = e.CountRedisKey(e.Framework.FrameworkName + ":server:*")

	d, err := json.Marshal(count)

	if err != nil {
		logrus.Error("V0GetCountK3SServer Error 1: ", err)
		return
	}

	logrus.Debug("HTTP GET V0GetCountK3SServer")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
