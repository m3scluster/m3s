package api

import (

	//"encoding/json"

	"encoding/json"

	"github.com/AVENTER-UG/mesos-m3s/controller"
	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	"github.com/gorilla/mux"

	//"io/ioutil"
	"net/http"

	"github.com/AVENTER-UG/mesos-m3s/mesos"
	"github.com/AVENTER-UG/mesos-m3s/redis"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
)

// API Service include all the current vars and global config
type API struct {
	Config            *cfg.Config
	Framework         *cfg.FrameworkConfig
	Mesos             mesos.Mesos
	Redis             *redis.Redis
	Kubernetes        *controller.Controller
	BootstrapProtocol string
	K3SAgentStatus    bool
}

// New will create a new API object
func New(cfg *cfg.Config, frm *cfg.FrameworkConfig) *API {
	e := &API{
		Config:            cfg,
		Framework:         frm,
		Mesos:             *mesos.New(cfg, frm),
		BootstrapProtocol: "http",
	}

	if e.Config.BootstrapSSLCrt != "" {
		e.BootstrapProtocol = "https"
	}

	return e
}

// Commands is the main function of this package
func (e *API) Commands() *mux.Router {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/api/m3s/versions", e.Versions).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/framework", e.V0FrameworkRemoveID).Methods("DELETE")
	rtr.HandleFunc("/api/m3s/v0/agent/scale/{count}", e.V0ScaleK3SAgent).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/agent/scale", e.V0GetCountK3SAgent).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/agent/cordon", e.V0CordonK3SAgent).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/agent/uncordon", e.V0UnCordonK3SAgent).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/server/scale/{count}", e.V0ScaleK3SServer).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/server/scale", e.V0GetCountK3SServer).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/server/config", e.V0GetKubeconfig).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/server/version", e.V0GetKubeVersion).Methods("GET")
	//rtr.HandleFunc("/api/m3s/v0/bootstrap/version", V0UpdateBootstrap).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/datastore/scale/{count}", e.V0ScaleDatastore).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/status/m3s", e.V0StatusM3s).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/status/k8s", e.V0StatusK8s).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/cluster/shutdown", e.V0ClusterShutdown).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/cluster/start", e.V0ClusterStart).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/cluster/restart", e.V0Restart).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/server/restart", e.V0Restart).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/agent/restart", e.V0Restart).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/ds/restart", e.V0Restart).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/capabilities", e.V0CapabilitiesDisclosure).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/{component}/{resource}/{value}", e.V0AdjustClusterResources).Methods("PUT")
	rtr.NotFoundHandler = http.HandlerFunc(e.NotFound)
	return rtr
}

// NotFound logs filenotfound messages
func (e *API) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	logrus.WithField("func", "api.NotFound").Error("404: " + r.RequestURI)
}

// Versions give out a list of Versions
func (e *API) Versions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "-")
	w.Write([]byte("/api/m3s/v0"))
}

// CheckAuth will check if the token is valid
func (e *API) CheckAuth(r *http.Request, w http.ResponseWriter) bool {
	// if no credentials are configured, then we dont have to check
	if e.Config.Credentials.Username == "" || e.Config.Credentials.Password == "" {
		return true
	}

	username, password, ok := r.BasicAuth()

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}

	if username == e.Config.Credentials.Username && password == e.Config.Credentials.Password {
		return true
	}

	w.WriteHeader(http.StatusUnauthorized)
	return false
}

// ErrorMessage will create a message json
func (e *API) ErrorMessage(number int, function string, msg string) []byte {
	var err cfg.ErrorMsg
	err.Function = function
	err.Number = number
	err.Message = msg

	data, _ := json.Marshal(err)
	return data
}
