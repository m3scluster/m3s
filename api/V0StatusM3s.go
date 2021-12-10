package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0StatusM3s gives out the current status of the M3s services
// example:
// curl -X GET 127.0.0.1:10000/v0/status/m3s -d 'JSON'
func V0StatusM3s(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	status := config.M3SStatus

	d, _ := json.Marshal(&status)

	logrus.Debug("HTTP GET V0StatusM3s: ", string(d))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
