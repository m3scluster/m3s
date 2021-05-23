package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"

	"github.com/AVENTER-UG/util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var MinVersion string

// Commands is the main function of this package
func Commands() *mux.Router {
	// Connect with database

	rtr := mux.NewRouter()
	rtr.HandleFunc("/versions", APIVersions).Methods("GET")
	rtr.HandleFunc("/api/k3s/v0/config", APIGetKubeConfig).Methods("GET")

	return rtr
}

// APIVersions give out a list of Versions
func APIVersions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "-")
	w.Write([]byte("/api/k3s/v0"))
}

// APIGetKubeConfig
func APIGetKubeConfig(w http.ResponseWriter, r *http.Request) {

	content, err := ioutil.ReadFile("/mnt/mesos/sandbox/kubeconfig.yaml")
	if err != nil {
		logrus.Error("Error reading file:", err)
		w.Write([]byte("Error reading kubeconfig.yaml"))
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Api-Service", "v0")

		w.Write(content)
	}

}

func main() {
	util.SetLogging("INFO", false, "GO-K3S-API")

	bind := flag.String("bind", "0.0.0.0", "The IP address to bind")
	port := flag.String("port", "10422", "The port to listen")

	logrus.Println("GO-K3S-API build"+MinVersion, *bind, *port)

	http.Handle("/", Commands())

	if err := http.ListenAndServe(*bind+":"+*port, nil); err != nil {
		logrus.Fatalln("ListenAndServe: ", err)
	}

}
