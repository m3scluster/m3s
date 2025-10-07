package api

import (
	"net/http"
	"strconv"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	"github.com/gorilla/mux"
)

// V0AdjustClusterResources will adjust the resource allocated to the k3s clusters
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/agent/memory/{value} -d 'JSON'
func (e *API) V0AdjustClusterResources(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0AdjustClusterResources").Debug("Call")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	if vars == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	component := vars["component"]
	resource := vars["value"]
	value, err := strconv.Atoi(vars["value"])
	if component == "" || resource == "" || err != nil {
		logrus.WithField("func", "api.V0AdjustClusterResources").Error("Error: Required Fields/Parameters are empty or wrong: ", err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	d := e.ErrorMessage(0, "V0AdjustClusterResources", "ok")

	switch component {
	case "ds":
		if resource == "cpus" {
			e.Config.DSCPU = float64(value)
		} else if resource == "cpusLimit" {
			e.Config.DSCPULimit = float64(value)
		} else if resource == "memory" {
			e.Config.DSMEM = float64(value)
		} else if resource == "memoryLimit" {
			e.Config.DSMEMLimit = float64(value)
		} else if resource == "disk" {
			e.Config.DSDISK = float64(value)
		} else if resource == "diskLimit" {
			e.Config.DSDISKLimit = float64(value)
		}
	case "server":
		if resource == "cpus" {
			e.Config.K3SServerCPU = float64(value)
		} else if resource == "cpusLimit" {
			e.Config.K3SServerCPULimit = float64(value)
		} else if resource == "memory" {
			e.Config.K3SServerMEM = float64(value)
		} else if resource == "memoryLimit" {
			e.Config.K3SServerMEMLimit = float64(value)
		} else if resource == "disk" {
			e.Config.K3SServerDISK = float64(value)
		} else if resource == "diskLimit" {
			e.Config.K3SServerDISKLimit = float64(value)
		}
	case "agent":
		if resource == "cpus" {
			e.Config.K3SAgentCPU = float64(value)
		} else if resource == "cpusLimit" {
			e.Config.K3SAgentCPULimit = float64(value)
		} else if resource == "memory" {
			e.Config.K3SAgentMEM = float64(value)
		} else if resource == "memoryLimit" {
			e.Config.K3SAgentMEMLimit = float64(value)
		} else if resource == "disk" {
			e.Config.K3SAgentDISK = float64(value)
		} else if resource == "diskLimit" {
			e.Config.K3SAgentDISKLimit = float64(value)
		}
	}

	e.Redis.SaveConfig(*e.Config)
	d = []byte("Resources Updated")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}
