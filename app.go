package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"mesos-k3s/api"
	"mesos-k3s/mesos"
	cfg "mesos-k3s/types"

	util "github.com/AVENTER-UG/util"
	"github.com/Showmax/go-fqdn"
	"github.com/sirupsen/logrus"
)

// MinVersion is the version number of this program
var MinVersion string

func main() {
	util.SetLogging(config.LogLevel, config.EnableSyslog, config.AppName)
	logrus.Println(config.AppName + " build " + MinVersion)

	hostname := fqdn.Get()
	listen := fmt.Sprintf(":%s", config.FrameworkPort)

	logrus.Info(hostname)

	failoverTimeout := 5000.0
	checkpoint := true
	webuiurl := fmt.Sprintf("http://%s%s", hostname, listen)

	config.FrameworkInfoFile = fmt.Sprintf("%s/%s", config.FrameworkInfoFilePath, "framework.json")
	config.CommandChan = make(chan cfg.Command, 100)
	config.Hostname = hostname
	config.Listen = listen

	config.State = map[string]cfg.State{}

	config.FrameworkInfo.User = config.FrameworkUser
	config.FrameworkInfo.Name = config.FrameworkName
	config.FrameworkInfo.Hostname = &hostname
	config.FrameworkInfo.WebUiURL = &webuiurl
	config.FrameworkInfo.FailoverTimeout = &failoverTimeout
	config.FrameworkInfo.Checkpoint = &checkpoint
	config.FrameworkInfo.Principal = &config.Principal
	config.FrameworkInfo.Role = &config.FrameworkRole
	//	config.FrameworkInfo.Capabilities = []mesosproto.FrameworkInfo_Capability{
	//		{Type: mesosproto.FrameworkInfo_Capability_RESERVATION_REFINEMENT},
	//	}

	// Load the old state if its exist
	frameworkJSON, err := ioutil.ReadFile(config.FrameworkInfoFile)
	if err == nil {
		json.Unmarshal([]byte(frameworkJSON), &config)
		mesos.Reconcile()
	}

	mesos.SetConfig(&config)
	api.SetConfig(&config)

	http.Handle("/", api.Commands())

	go func() {
		http.ListenAndServe(listen, nil)
	}()
	logrus.Fatal(mesos.Subscribe())
}
