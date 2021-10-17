package api

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// V0GetKubeVersion will return the kubernetes Version
// example:
// curl -X GET 127.0.0.1:10000/v0/server/version'
func V0GetKubeVersion(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("HTTP GET V0GetKubeVersion")

	auth := CheckAuth(r, w)

	if !auth {
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+config.M3SBootstrapServerHostname+":"+strconv.Itoa(config.M3SBootstrapServerPort)+"/api/k3s/v0/version", nil)
	req.Close = true
	res, err := client.Do(req)

	if err != nil {
		logrus.Error("V0GetKubeVersion: Error 1: ", err, res)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.Error("V0GetKubeVersion: Error Status is not 200")
		return
	}

	content, err := ioutil.ReadAll(res.Body)

	if err != nil {
		logrus.Error("V0GetKubeVersion: Error 2: ", err, res)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write([]byte(content))
}
