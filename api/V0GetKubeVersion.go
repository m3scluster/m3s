package api

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// V0GetKubeVersion will return the kubernetes Version
// example:
// curl -X GET 127.0.0.1:10000/v0/server/version'
func (e *API) V0GetKubeVersion(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0GetKubeVersion").Debug("Call")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	client := &http.Client{}
	// #nosec G402
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: e.Config.SkipSSL},
	}
	req, _ := http.NewRequest("GET", e.BootstrapProtocol+"://"+e.Config.K3SServerHostname+":"+strconv.Itoa(e.Config.K3SServerContainerPort)+"/api/m3s/bootstrap/v0/version", nil)
	req.SetBasicAuth(e.Config.BootstrapCredentials.Username, e.Config.BootstrapCredentials.Password)
	req.Close = true
	res, err := client.Do(req)

	if err != nil {
		logrus.WithField("func", "V0GetKubeVersion").Error(err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.WithField("func", "V0GetKubeVersion").Error("Response Code: ", res.StatusCode)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	content, err := io.ReadAll(res.Body)

	if err != nil {
		logrus.WithField("func", "V0GetKubeVersion").Error(err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	err = json.Unmarshal(content, &e.Config.Version)
	if err != nil {
		logrus.WithField("func", "V0GetKubeVersion").Error(err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	d, err := json.Marshal(&e.Config.Version)
	if err != nil {
		logrus.Error("V0GetKubeVersion: Error 4 ", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}
