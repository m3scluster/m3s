package api

import (
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

// V0StatusContainer will return the status of the taskID
// example:
// curl -X GET 127.0.0.1:10000/v0/k3s/config'
func V0GetKubeconfig(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("HTTP GET V0GetKubeconfig")

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+config.K3SServerAPIHostname+":"+config.K3SServerAPIPort+"/api/k3s/v0/config", nil)
	res, err := client.Do(req)

	if err != nil {
		logrus.Error("GetKubeConfig: Error 1: ", err, res)
		return
	}

	if res.StatusCode != 200 {
		logrus.Error("GetKubeConfig: Error Status is not 200")
		return
	}

	content, err := ioutil.ReadAll(res.Body)

	if err != nil {
		logrus.Error("GetKubeConfig: Error 2: ", err, res)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(content)

}
