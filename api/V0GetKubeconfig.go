package api

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// V0StatusContainer will return the status of the taskID
// example:
// curl -X GET 127.0.0.1:10000/v0/k3s/config'
func V0GetKubeconfig(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("HTTP GET V0GetKubeconfig")

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+config.K3SServerAPIHostname+":"+strconv.Itoa(config.K3SServerAPIPort)+"/api/k3s/v0/config", nil)
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

	// replace the localhost server string with the mesos agent hostname and dynamic port
	destUrl := config.K3SServerAPIHostname + ":" + strconv.Itoa(config.K3SServerPort)
	kubconf := strings.Replace(string(content), "127.0.0.1:6443", destUrl, -1)

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write([]byte(kubconf))

}
