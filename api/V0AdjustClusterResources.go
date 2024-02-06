package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	"github.com/gorilla/mux"
)

// V0AdjustClusterResources will adjust the resource allocated to the k3s clusters
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/agent/memory/{value} -d 'JSON'
func (e *API) V0AdjustClusterResources(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0AdjustClusterResources").Debug("Call")

	vars := mux.Vars(r)

	pathFragments := strings.Split(r.URL.Path, "/")

	component := pathFragments[4]
	resource := pathFragments[5]

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if vars == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	d := e.ErrorMessage(0, "V0AdjustClusterResources", "ok")

	if vars["value"] != "" {
		value, err := strconv.Atoi(vars["value"])
		if err != nil {
			logrus.WithField("func", "api.V0AdjustClusterResources").Error("Error: ", err.Error())
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}

		fmt.Println(resource)

		switch component {
		case "ds":
			if resource == "cpus" {
				e.Config.DSCPU = float64(value)
			} else if resource == "memory" {
				e.Config.DSMEM = float64(value)
			}
		case "server":
			if resource == "cpus" {
				e.Config.K3SServerCPU = float64(value)
			} else if resource == "memory" {
				e.Config.K3SServerMEM = float64(value)
			}
		case "agent":
			if resource == "cpus" {
				e.Config.K3SAgentCPU = float64(value)
			} else if resource == "memory" {
				e.Config.K3SAgentMEM = float64(value)
			}
		}

		e.Redis.SaveConfig(*e.Config)
		d = []byte("Resources Updated")

	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}
