package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// V0GetKubeVersion will return the kubernetes Version
// example:
// curl -X GET 127.0.0.1:10000/v0/server/version'
func (e *API) V0GetKubeVersion(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("HTTP GET V0GetKubeVersion")

	auth := e.CheckAuth(r, w)

	if !auth {
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+e.Config.M3SBootstrapServerHostname+":"+strconv.Itoa(e.Config.M3SBootstrapServerPort)+"/api/m3s/bootstrap/v0/version", nil)
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

	err = json.Unmarshal(content, &e.Config.Version)
	if err != nil {
		logrus.Error("V0GetKubeVersion: Error 3 ", err)
		return
	}

	d, err := json.Marshal(&e.Config.Version)
	if err != nil {
		logrus.Error("V0GetKubeVersion: Error 4 ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}
