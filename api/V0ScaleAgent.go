package api

import (
	"net/http"
	"strconv"
	"time"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	"github.com/gorilla/mux"
)

// V0ScaleK3SAgent will scale the k3s agent service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/agent/scale/{count of instances} -d 'JSON'
func (e *API) V0ScaleK3SAgent(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0ScaleK3SAgent").Debug("Call")

	vars := mux.Vars(r)

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if vars == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	d := e.ErrorMessage(0, "V0ScaleK3SAgent", "ok")

	if vars["count"] != "" {
		count, err := strconv.Atoi(vars["count"])
		if err != nil {
			logrus.WithField("func", "api.V0ScaleK3SAgent").Error("Error: ", err.Error())
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		d = e.scaleAgent(count)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")
	w.Write(d)
}

// agentStop will scale down the agent
func (e *API) agentStop() {
	logrus.WithField("func", "api.Sgenttop").Debug("Shutdown Agent")

	// Save current amount of services for the case of restart but only
	// if the amount is not 0
	if e.Config.K3SAgentMax != 0 {
		e.Config.K3SAgentMaxRestore = e.Config.K3SAgentMax
	}

	e.scaleAgent(0)
	e.Mesos.SuppressFramework()
}

// agentStart will scale up the agents
func (e *API) agentStart() {
	logrus.WithField("func", "api.agentStart").Debug("Start Agent")
	e.scaleAgent(e.Config.K3SAgentMaxRestore)
}

// AgentRestart will scale down all K8 agents and scale up again.
func (e *API) AgentRestart() {
	e.agentStop()

	ticker := time.NewTicker(e.Config.EventLoopTime)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		if e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", "") == 0 {
			logrus.WithField("func", "api.AgentRestart").Debug("All agents down")
			goto start
		}
	}

start:
	ticker.Stop()
	e.agentStart()
}

// scaleAgent - can scale up and down the K8 worker nodes
func (e *API) scaleAgent(count int) []byte {
	r := e.scale(count, e.Config.K3SAgentMax, "agent")
	e.Config.K3SAgentMax = count
	return r
}
