package api

import (

	//"encoding/json"

	"github.com/gorilla/mux"
	//"io/ioutil"
	"net/http"

	cfg "mesos-k3s/types"
)

// Service include all the current vars and global config
var config *cfg.Config

// SetConfig set the global config
func SetConfig(cfg *cfg.Config) {
	config = cfg
}

// Commands is the main function of this package
func Commands() *mux.Router {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/v0/container/{taskID}", V0StatusContainer).Methods("GET")
	rtr.HandleFunc("/v0/agent/scale/{count}", V0ScaleK3SAgent).Methods("GET")
	rtr.HandleFunc("/v0/agent/reflate", V0ReflateK3SAgent).Methods("GET")
	rtr.HandleFunc("/v0/server/scale/{count}", V0ScaleK3S).Methods("GET")
	rtr.HandleFunc("/v0/server/reflate", V0ReflateK3S).Methods("GET")
	rtr.HandleFunc("/v0/etcd/scale/{count}", V0ScaleEtcd).Methods("GET")
	rtr.HandleFunc("/v0/task/kill/{id}", V0KillTask).Methods("GET")

	return rtr
}

// CheckAuth will check if the token is valid
func CheckAuth(r *http.Request, w http.ResponseWriter) bool {
	// if no credentials are configured, then we dont have to check
	if config.Credentials.Username == "" || config.Credentials.Password == "" {
		return true
	}

	username, password, ok := r.BasicAuth()

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}

	if username == config.Credentials.Username && password == config.Credentials.Password {
		w.WriteHeader(http.StatusOK)
		return true
	}

	w.WriteHeader(http.StatusUnauthorized)
	return false
}
