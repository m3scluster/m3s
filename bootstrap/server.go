package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/AVENTER-UG/util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	chkVersion "github.com/hashicorp/go-version"
)

// BuildVersion of m3s
var BuildVersion string

// GitVersion is the revision and commit number
var GitVersion string

// VersionURL is the URL of the .version.json file
var VersionURL string

// DashboardInstalled is true if the dashboard is already installed
var DashboardInstalled bool

// TraefikDashboardInstalled is true if the traefik dashboard is installed
var TraefikDashboardInstalled bool

// Commands is the main function of this package
func Commands() *mux.Router {
	// Connect with database

	rtr := mux.NewRouter()
	rtr.HandleFunc("/api/m3s/bootstrap/versions", APIVersions).Methods("GET")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/update", APIUpdate).Methods("PUT")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/status", APIHealth).Methods("GET")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/config", APIGetKubeConfig).Methods("GET")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/version", APIGetKubeVersion).Methods("GET")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/status?verbose", APIStatus).Methods("GET")

	return rtr
}

// APIVersions give out a list of Versions
func APIVersions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "-")
	w.Write([]byte("/api/m3s/bootstrap/v0"))
}

// APIUpdate do a update of the bootstrap server
func APIUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	// check first if there is a update
	logrus.Info("Get version file: ", VersionURL)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", VersionURL, nil)
	req.Close = true
	res, err := client.Do(req)

	if err != nil {
		logrus.Error("APIUpdate: Error 1: ", err, res)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Error("APIUpdate: Error Status is not 200")
		return
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Error("APIUpdate: Error 2: ", err, res)
		return
	}

	var version cfg.M3SVersion
	err = json.Unmarshal(body, &version)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Error("APIUpdate: Error 3: ", err, body)
		return
	}

	// check if the current Version diffs to the online version. If yes, then start the update.
	newVersion, _ := chkVersion.NewVersion(version.BootstrapVersion.GitVersion)
	currentVersion, _ := chkVersion.NewVersion(GitVersion)

	logrus.Info("APIUpdate newVersion: ", newVersion)
	logrus.Info("APIUpdate currentVersion: ", currentVersion)

	if currentVersion.LessThan(newVersion) {
		w.Write([]byte("Start bootstrap server update"))
		logrus.Info("Start update")
		// #nosec: G204
		stdout, err := exec.Command("/mnt/mesos/sandbox/update", strconv.Itoa(os.Getpid())).Output()
		if err != nil {
			logrus.Error("Do update", err, stdout)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.Write([]byte("No update for the bootstrap server"))
	}
}

// APIGetKubeConfig get out the kubernetes config file
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

// APIGetKubeVersion get out the kubernetes version number
func APIGetKubeVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	stdout, err := exec.Command("/mnt/mesos/sandbox/kubectl", "version", "-o=json").Output()
	if err != nil {
		logrus.Error("Get Kubernetes Version: ", err, stdout)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(stdout)
}

// APIHealth give out the status of the kubernetes server
func APIHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	logrus.Debug("Health Check")

	// check if the kubernetes server is working
	stdout, err := exec.Command("/mnt/mesos/sandbox/kubectl", "get", "--raw=/livez/ping").Output()

	if err != nil {
		logrus.Error("Health to Kubernetes Server: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if string(stdout) == "ok" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

		// if kubernetes server is running and the dashboard is not installed, then do it
		if !DashboardInstalled {
			deployDashboard()
		}
		// if kubernetes server is running and the traefik dashboard is not installed, then do it
		if !TraefikDashboardInstalled {
			deployTraefikDashboard()
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// deployDashboard will deploy the kubernetes dashboard
// if the server is in the running state
func deployDashboard() {
	err := exec.Command("/mnt/mesos/sandbox/kubectl", "apply", "-f", "/mnt/mesos/sandbox/dashboard.yaml").Run()
	logrus.Info("Install Kubernetes Dashboard")

	if err != nil {
		logrus.Error("Install Kubernetes Dashboard: ", err)
		return
	}

	err = exec.Command("/mnt/mesos/sandbox/kubectl", "apply", "-f", "/mnt/mesos/sandbox/dashboard_auth.yaml").Run()

	if err != nil {
		logrus.Error("Install Kubernetes Dashboard Auth: ", err)
		return
	}

	logrus.Info("Install Kubernetes Dashboard: Done")
	DashboardInstalled = true
}

// deployTraefikDashboard will deploy the traefik dashboard
// if the server is in the running state
func deployTraefikDashboard() {
	err := exec.Command("/mnt/mesos/sandbox/kubectl", "apply", "-f", "/mnt/mesos/sandbox/dashboard_traefik.yaml").Run()
	logrus.Info("Install Traefik Dashboard")

	if err != nil {
		logrus.Error("Install Traefik Dashboard: ", err)
		return
	}

	logrus.Info("Install Traefik Dashboard: Done")
	TraefikDashboardInstalled = true
}

// APIStatus give out the status of the kubernetes server
func APIStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	logrus.Debug("Status Information")

	// check if the kubernetes server is working
	stdout, err := exec.Command("/mnt/mesos/sandbox/kubectl", "get", "--raw='/readyz?verbose'").Output()

	if err != nil {
		logrus.Error("Health to Kubernetes Server: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(stdout)
}

func main() {
	// Prints out current version
	var version bool
	flag.BoolVar(&version, "v", false, "Prints current version")
	flag.Parse()
	if version {
		fmt.Print(GitVersion)
		return
	}

	util.SetLogging("INFO", false, "GO-K3S-API")

	bind := flag.String("bind", "0.0.0.0", "The IP address to bind")
	port := flag.String("port", "10422", "The port to listen")

	logrus.Println("GO-K3S-API build "+BuildVersion+" git "+GitVersion+" ", *bind, *port)

	DashboardInstalled = false

	http.Handle("/", Commands())

	if err := http.ListenAndServe(*bind+":"+*port, nil); err != nil {
		logrus.Fatalln("ListenAndServe: ", err)
	}
}
