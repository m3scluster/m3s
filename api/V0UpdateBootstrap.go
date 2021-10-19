package api

import (
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// V0UpdateBootstrap will update the bootstrap server in the k8 manger container
// example:
// curl -X GET 127.0.0.1:10000/v0/bootstrap/update'
func V0UpdateBootstrap(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("HTTP PUT V0UpdateBootstrap")

	auth := CheckAuth(r, w)

	if !auth {
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+config.M3SBootstrapServerHostname+":"+strconv.Itoa(config.M3SBootstrapServerPort)+"/update", nil)
	req.Close = true
	res, err := client.Do(req)

	if err != nil {
		logrus.Error("V0UpdateBootstrap: Error 1: ", err, res)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.Error("V0UpdateBootstrap: Error Status is not 200")
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write([]byte("ok"))
}
