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

// Service include all the current vars and global config
var config *cfg.Config
var framework *mesosutil.FrameworkConfig

// SetConfig set the global config
func SetConfig(cfg *cfg.Config, frm *mesosutil.FrameworkConfig) {
	config = cfg
	framework = frm
}

// Commands is the main function of this package
func Commands() *mux.Router {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/v0/agent/scale/{count}", V0ScaleK3SAgent).Methods("GET")
	rtr.HandleFunc("/v0/agent/scale", V0GetCountK3SAgent).Methods("GET")
	rtr.HandleFunc("/v0/server/config", V0GetKubeconfig).Methods("GET")
	rtr.HandleFunc("/v0/server/version", V0GetKubeVersion).Methods("GET")
	rtr.HandleFunc("/v0/bootstrap/update", V0UpdateBootstrap).Methods("PUT")
	//	rtr.HandleFunc("/v0/bootstrap/version", V0UpdateBootstrap).Methods("PUT")
	rtr.HandleFunc("/v0/etcd/scale/{count}", V0ScaleEtcd).Methods("GET")
	rtr.HandleFunc("/v0/status/m3s", V0StatusM3s).Methods("GET")
	rtr.HandleFunc("/v0/status/k8s", V0StatusK8s).Methods("GET")

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
		return true
	}

	w.WriteHeader(http.StatusUnauthorized)
	return false
}

// ErrorMessage will create a message json
func ErrorMessage(number int, function string, msg string) []byte {
	var err cfg.ErrorMsg
	err.Function = function
	err.Number = number
	err.Message = msg

	data, _ := json.Marshal(err)
	return []byte(data)

}
