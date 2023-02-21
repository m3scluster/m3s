package api

import (
	"crypto/tls"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// V0GetKubeconfig will return the kubernetes config file
// example:
// curl -X GET 127.0.0.1:10000/api/m3s/v0/server/config"
func (e *API) V0GetKubeconfig(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0GetKubeconfig").Debug("Call")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	client := &http.Client{}
	// #nosec G402
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: e.Config.SkipSSL},
	}
	req, _ := http.NewRequest("GET", e.BootstrapProtocol+"://"+e.Config.K3SServerHostname+":"+strconv.Itoa(e.Config.K3SServerContainerPort)+"/api/m3s/bootstrap/v0/config", nil)
	req.SetBasicAuth(e.Config.BootstrapCredentials.Username, e.Config.BootstrapCredentials.Password)
	req.Close = true
	res, err := client.Do(req)

	if err != nil {
		logrus.WithField("func", "V0GetKubeconfig").Error(err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.WithField("func", "V0GetKubeconfig").Error("Response Code: ", res.StatusCode)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	content, err := io.ReadAll(res.Body)

	if err != nil {
		logrus.WithField("func", "V0GetKubeconfig").Error(err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// replace the localhost server string with the mesos agent hostname and dynamic port
	destURL := e.Config.K3SServerHostname + ":" + strconv.Itoa(e.Config.K3SServerPort)
	kubconf := strings.Replace(string(content), "127.0.0.1:6443", destURL, -1)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write([]byte(kubconf))
}
