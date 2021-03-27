package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// V0StatusContainer will return the status of the taskID
// example:
// curl -X GET 127.0.0.1:10000/v0/container/{taskID} -d 'JSON'
func V0StatusContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	status := config.State[vars["taskID"]]

	d, _ := json.Marshal(&status)

	logrus.Debug("HTTP GET V0StatusContainer: ", string(d))

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
