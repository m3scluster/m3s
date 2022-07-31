package api

import (

	//"encoding/json"

	"encoding/json"

	"github.com/gorilla/mux"
	//"io/ioutil"
	"net/http"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	mesosutil "github.com/AVENTER-UG/mesos-util"
)

// API Service include all the current vars and global config
type API struct {
	Config    *cfg.Config
	Framework *mesosutil.FrameworkConfig
	Redis     Redis
}

// New will create a new API object
func New(cfg *cfg.Config, frm *mesosutil.FrameworkConfig) *API {
	e := &API{
		Config:    cfg,
		Framework: frm,
	}

	return e
}

// Commands is the main function of this package
func (e *API) Commands() *mux.Router {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/api/m3s/versions", e.Versions).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/agent/scale/{count}", e.V0ScaleK3SAgent).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/agent/scale", e.V0GetCountK3SAgent).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/server/scale/{count}", e.V0ScaleK3SServer).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/server/scale", e.V0GetCountK3SServer).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/server/config", e.V0GetKubeconfig).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/server/version", e.V0GetKubeVersion).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/bootstrap/update", e.V0UpdateBootstrap).Methods("PUT")
	//rtr.HandleFunc("/api/m3s/v0/bootstrap/version", V0UpdateBootstrap).Methods("PUT")
	rtr.HandleFunc("/api/m3s/v0/datastore/scale/{count}", e.V0ScaleDatastore).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/status/m3s", e.V0StatusM3s).Methods("GET")
	rtr.HandleFunc("/api/m3s/v0/status/k8s", e.V0StatusK8s).Methods("GET")

	return rtr
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
