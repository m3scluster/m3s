package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/AVENTER-UG/mesos-m3s/api"
	"github.com/AVENTER-UG/mesos-m3s/mesos"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	mesosutil "github.com/AVENTER-UG/mesos-util"

	util "github.com/AVENTER-UG/util"
	"github.com/sirupsen/logrus"
)

// BuildVersion of m3s
var BuildVersion string

// GitVersion is the revision and commit number
var GitVersion string

// convert Base64 Encodes PEM Certificate to tls object
func decodeBase64Cert(pemCert string) []byte {
	sslPem, err := base64.URLEncoding.DecodeString(pemCert)
	if err != nil {
		logrus.Fatal("Error decoding SSL PEM from Base64: ", err.Error())
	}
	return sslPem
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

	util.SetLogging(config.LogLevel, config.EnableSyslog, config.AppName)
	logrus.Println(config.AppName + " build " + BuildVersion + " git " + GitVersion)

	listen := fmt.Sprintf(":%s", framework.FrameworkPort)

	failoverTimeout := 5000.0
	checkpoint := true
	webuiurl := fmt.Sprintf("http://%s%s", framework.FrameworkHostname, listen)
	if config.SSLCrt != "" && config.SSLKey != "" {
		webuiurl = fmt.Sprintf("https://%s%s", framework.FrameworkHostname, listen)
	}

	framework.CommandChan = make(chan mesosutil.Command, 100)
	config.Hostname = framework.FrameworkHostname
	config.Listen = listen
	config.Suppress = false

	framework.State = map[string]mesosutil.State{}

	framework.FrameworkInfo.User = framework.FrameworkUser
	framework.FrameworkInfo.Name = framework.FrameworkName
	framework.FrameworkInfo.WebUiURL = &webuiurl
	framework.FrameworkInfo.FailoverTimeout = &failoverTimeout
	framework.FrameworkInfo.Checkpoint = &checkpoint
	framework.FrameworkInfo.Principal = &config.Principal
	framework.FrameworkInfo.Role = &framework.FrameworkRole

	//	config.FrameworkInfo.Capabilities = []mesosproto.FrameworkInfo_Capability{
	//		{Type: mesosproto.FrameworkInfo_Capability_RESERVATION_REFINEMENT},
	//	}

	mesosutil.SetConfig(&framework)
	mesos.SetConfig(&config, &framework)
	api.SetConfig(&config, &framework)

	api.ConnectRedis()

	// load framework state from DB
	key := api.GetRedisKey(framework.FrameworkName + ":framework")
	if key != "" {
		json.Unmarshal([]byte(key), &framework)
	}

	// restore variable data from the old config
	key = api.GetRedisKey(framework.FrameworkName + ":framework_config")
	if key != "" {
		var oldconfig cfg.Config
		json.Unmarshal([]byte(key), &oldconfig)
		config.M3SBootstrapServerPort = oldconfig.M3SBootstrapServerPort
		config.M3SBootstrapServerHostname = oldconfig.M3SBootstrapServerHostname
		config.K3SServerPort = oldconfig.K3SServerPort
		config.K3SServerURL = oldconfig.K3SServerURL
		config.K3SAgentMax = oldconfig.K3SAgentMax
		config.ETCDMax = oldconfig.ETCDMax

		api.SaveConfig()
	}

	// set current m3s version
	config.Version.M3SVersion.GitVersion = GitVersion
	config.Version.M3SVersion.BuildDate = BuildVersion

	// The Hostname should ever be set after reading the state file.
	framework.FrameworkInfo.Hostname = &framework.FrameworkHostname

	server := &http.Server{
		Addr:              listen,
		Handler:           api.Commands(),
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequestClientCert,
			MinVersion: tls.VersionTLS12,
		},
	}

	if config.SSLCrt != "" && config.SSLKey != "" {
		logrus.Debug("Enable TLS")
		crt := decodeBase64Cert(config.SSLCrt)
		key := decodeBase64Cert(config.SSLKey)
		certs, err := tls.X509KeyPair(crt, key)
		if err != nil {
			logrus.Fatal("TLS Server Error: ", err.Error())
		}
		server.TLSConfig.Certificates = []tls.Certificate{certs}
	}

	go func() {
		if config.SSLCrt != "" && config.SSLKey != "" {
			server.ListenAndServeTLS("", "")
		} else {
			server.ListenAndServe()
		}
	}()
	logrus.Fatal(mesos.Subscribe())
	config.RedisClient.Close()
}
