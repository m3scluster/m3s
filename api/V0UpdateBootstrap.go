package api

import (
	"crypto/tls"
	"io"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// V0UpdateBootstrap will update the bootstrap server in the k8 manger container
// example:
// curl -X GET 127.0.0.1:10000/v0/bootstrap/update'
func (e *API) V0UpdateBootstrap(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0UpdateBootstrap").Debug("Call")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	client := &http.Client{}
	// #nosec G402
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: e.Config.SkipSSL},
	}
	req, _ := http.NewRequest("PUT", e.BootstrapProtocol+"://"+e.Config.K3SServerHostname+":"+strconv.Itoa(e.Config.K3SServerContainerPort)+"/api/m3s/bootstrap/v0/update", nil)
	req.SetBasicAuth(e.Config.BootstrapCredentials.Username, e.Config.BootstrapCredentials.Password)
	req.Close = true
	res, err := client.Do(req)

	if err != nil {
		logrus.WithField("func", "V0UpdateBootstrap").Error(err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.WithField("func", "V0UpdateBootstrap").Error("Response Code: ", res.StatusCode)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	content, err := io.ReadAll(res.Body)

	if err != nil {
		logrus.WithField("func", "V0UpdateBootstrap").Error(err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(content)
}
