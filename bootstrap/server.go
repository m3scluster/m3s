package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/AVENTER-UG/util/util"
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

var config cfg.Config

// convert Base64 Encodes PEM Certificate to tls object
func decodeBase64Cert(pemCert string) []byte {
	sslPem, err := base64.URLEncoding.DecodeString(pemCert)
	if err != nil {
		logrus.Fatal("Error decoding SSL PEM from Base64: ", err.Error())
	}
	return sslPem
}

// Commands is the main function of this package
func Commands() *mux.Router {
	// Connect with database

	rtr := mux.NewRouter()
	rtr.HandleFunc("/api/m3s/bootstrap/versions", APIVersions).Methods("GET")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/update", APIUpdate).Methods("PUT")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/status", APIHealth).Methods("GET")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/config", APIGetKubeConfig).Methods("GET")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/version", APIGetKubeVersion).Methods("GET")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/clean", APICleanupNotRead).Methods("GET")
	rtr.HandleFunc("/api/m3s/bootstrap/v0/status?verbose", APIStatus).Methods("GET")
	rtr.NotFoundHandler = http.HandlerFunc(NotFound)

	return rtr
}

// NotFound logs filenotfound messages
func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	logrus.WithField("func", "bootstrap.NotFound").Error("404: " + r.RequestURI)
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

	if !CheckAuth(r, w) {
		return
	}

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

	body, err := io.ReadAll(res.Body)

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
	if !CheckAuth(r, w) {
		return
	}

	content, err := os.ReadFile("/mnt/mesos/sandbox/kubeconfig.yaml")
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

	if !CheckAuth(r, w) {
		return
	}

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

	if !CheckAuth(r, w) {
		return
	}

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

	if !CheckAuth(r, w) {
		return
	}

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

// APICleanupNotRead -  cleanup notready nodes
func APICleanupNotRead(w http.ResponseWriter, r *http.Request) {
	if !CheckAuth(r, w) {
		return
	}

	cmd := "/mnt/mesos/sandbox/kubectl delete node $(kubectl get nodes | grep NotReady | awk '{print $1;}')"
	err := exec.Command("bash", "-c", cmd).Run()
	logrus.Info("Cleanup notready nodes")

	if err != nil {
		logrus.Error("Cleanup notready nodes: ", err)
		return
	}

	logrus.Info("Cleanup notready nodes: Done")
}

// CheckAuth will check if the token is valid
func CheckAuth(r *http.Request, w http.ResponseWriter) bool {
	// if no credentials are configured, then we dont have to check
	if config.BootstrapCredentials.Username == "" || config.BootstrapCredentials.Password == "" {
		return true
	}

	username, password, ok := r.BasicAuth()

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}

	if username == config.BootstrapCredentials.Username && password == config.BootstrapCredentials.Password {
		return true
	}

	w.WriteHeader(http.StatusUnauthorized)
	return false
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

	config.BootstrapCredentials.Username = util.Getenv("BOOTSTRAP_AUTH_USERNAME", "")
	config.BootstrapCredentials.Password = util.Getenv("BOOTSTRAP_AUTH_PASSWORD", "")
	config.BootstrapSSLKey = util.Getenv("BOOTSTRAP_SSL_KEY_BASE64", "")
	config.BootstrapSSLCrt = util.Getenv("BOOTSTRAP_SSL_CRT_BASE64", "")

	logrus.Println("GO-K3S-API build "+BuildVersion+" git "+GitVersion+" ", *bind, *port)

	DashboardInstalled = false

	listen := fmt.Sprintf(":%s", *port)

	server := &http.Server{
		Addr:              listen,
		Handler:           Commands(),
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequestClientCert,
			MinVersion: tls.VersionTLS12,
		},
	}

	if config.BootstrapSSLCrt != "" && config.BootstrapSSLKey != "" {
		logrus.Debug("Enable TLS")
		crt := decodeBase64Cert(config.BootstrapSSLCrt)
		key := decodeBase64Cert(config.BootstrapSSLKey)
		certs, err := tls.X509KeyPair(crt, key)
		if err != nil {
			logrus.Fatal("TLS Server Error: ", err.Error())
		}
		server.TLSConfig.Certificates = []tls.Certificate{certs}
	}

	if config.BootstrapSSLCrt != "" && config.BootstrapSSLKey != "" {
		server.ListenAndServeTLS("", "")
	} else {
		server.ListenAndServe()
	}
}
