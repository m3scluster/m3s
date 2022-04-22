package api

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0StatusK8s gives out the current status of the K8s services
// example:
// curl -X GET 127.0.0.1:10000/v0/status/k8s
func V0StatusK8s(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("HTTP GET V0StatusK8s ")
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+config.M3SBootstrapServerHostname+":"+strconv.Itoa(config.M3SBootstrapServerPort)+"/api/m3s/bootstrap/v0/status?verbose", nil)
	req.Close = true
	res, err := client.Do(req)

	if err != nil {
		logrus.Error("StatusK8s: Error 1: ", err, res)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.Error("StatusK8s: Error Status is not 200")
		return
	}

	content, err := ioutil.ReadAll(res.Body)

	if err != nil {
		logrus.Error("StatusK8s: Error 2: ", err, res)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(content)
}
