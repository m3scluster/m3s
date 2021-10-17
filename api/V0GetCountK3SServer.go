package api

import (
	"encoding/json"
	"net/http"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	"github.com/sirupsen/logrus"
)

// V0GetCountK3S will write out the current count of Kubernetes Server and the expected one.
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/server/scale'
func V0GetCountK3S(w http.ResponseWriter, r *http.Request) {
	auth := CheckAuth(r, w)

	if !auth {
		return
	}

	var count cfg.Count

	count.Scale = config.K3SServerMax
	count.Running = config.K3SServerCount

	d, err := json.Marshal(count)

	if err != nil {
		logrus.Error("V0GetCountK3S Error 1: ", err)
		return
	}

	logrus.Debug("HTTP GET V0GetCountK3S")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
