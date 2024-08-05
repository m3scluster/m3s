package api

import (
	"encoding/json"
	"net/http"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// V0CapabilitiesDisclosure gives out the current capabilities supported in this m3s framework version.
// Helpful when you may have multiple versions running in your environment and need feature flagging on downstream applications.
// example:
// curl -X GET 127.0.0.1:10000/v0/capabilities
func (e *API) V0CapabilitiesDisclosure(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0CapabilitiesDisclosure").Debug("Call")

	if !e.CheckAuth(r, w) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	capabilities := []string{}
	capabilities = append(capabilities, "cluster-restart, /api/m3s/v0/cluster/restart, PUT")
	capabilities = append(capabilities, "server-restart, /api/m3s/v0/server/restart, PUT")
	capabilities = append(capabilities, "agent-restart, /api/m3s/v0/agent/restart, PUT")
	capabilities = append(capabilities, "ds-restart, /api/m3s/v0/ds/restart, PUT")
	capabilities = append(capabilities, "cluster-start, /api/m3s/v0/cluster/start, PUT")
	capabilities = append(capabilities, "cluster-shutdown, /api/m3s/v0/cluster/shutdown, PUT")
	capabilities = append(capabilities, "k3s-count-agents, /api/m3s/v0/agent/scale, GET")
	capabilities = append(capabilities, "k3s-count-servers, /api/m3s/v0/server/scale, GET")
	capabilities = append(capabilities, "get-kubeconfig, /api/m3s/v0/server/config, GET")
	capabilities = append(capabilities, "get-kubeversion, /api/m3s/v0/server/version, GET")
	capabilities = append(capabilities, "scale-agent, /api/m3s/v0/agent/scale/{count}, GET")
	capabilities = append(capabilities, "scale-server, /api/m3s/v0/server/scale/{count}, GET")
	capabilities = append(capabilities, "scale-datastore, /api/m3s/v0/datastore/scale/{count}, GET")
	capabilities = append(capabilities, "status-k3s, /api/m3s/v0/status/k3s, GET")
	capabilities = append(capabilities, "status-m3s, /api/m3s/v0/status/m3s, GET")
	capabilities = append(capabilities, "adjust-server-cpus, /api/m3s/v0/server/cps/{cpus}, PUT")
	capabilities = append(capabilities, "adjust-agent-cpus, /api/m3s/v0/agent/cps/{cpus}, PUT")
	capabilities = append(capabilities, "adjust-ds-cpus, /api/m3s/v0/ds/cps/{cpus}, PUT")
	capabilities = append(capabilities, "adjust-server-memory, /api/m3s/v0/server/memory/{memory}, PUT")
	capabilities = append(capabilities, "adjust-agent-memory, /api/m3s/v0/agent/memory/{memory}, PUT")
	capabilities = append(capabilities, "adjust-ds-memory, /api/m3s/v0/ds/memory/{memory}, PUT")
	capabilities = append(capabilities, "cordon-agent-nodes, /api/m3s/v0/agent/cordon, PUT")
	capabilities = append(capabilities, "uncordon-agent-nodes, /api/m3s/v0/agent/uncordon, PUT")

	response, _ := json.Marshal(capabilities)
	w.Write(response)
}
