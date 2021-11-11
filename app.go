package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/AVENTER-UG/mesos-m3s/api"
	"github.com/AVENTER-UG/mesos-m3s/mesos"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"

	util "github.com/AVENTER-UG/util"
	"github.com/sirupsen/logrus"
)

// MinVersion is the version number of this program
var MinVersion string

func main() {
	util.SetLogging(config.LogLevel, config.EnableSyslog, config.AppName)
	logrus.Println(config.AppName + " build " + MinVersion)

	listen := fmt.Sprintf(":%s", config.FrameworkPort)

	failoverTimeout := 5000.0
	checkpoint := true
	webuiurl := fmt.Sprintf("http://%s%s", config.FrameworkHostname, listen)

	config.FrameworkInfoFile = fmt.Sprintf("%s/%s", config.FrameworkInfoFilePath, "framework.json")
	config.CommandChan = make(chan cfg.Command, 100)
	config.Hostname = config.FrameworkHostname
	config.Listen = listen

	config.State = map[string]cfg.State{}

	config.FrameworkInfo.User = config.FrameworkUser
	config.FrameworkInfo.Name = config.FrameworkName
	config.FrameworkInfo.WebUiURL = &webuiurl
	config.FrameworkInfo.FailoverTimeout = &failoverTimeout
	config.FrameworkInfo.Checkpoint = &checkpoint
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
	// The Hostname should ever be set after reading the state file.
	config.FrameworkInfo.Hostname = &config.FrameworkHostname

	mesos.SetConfig(&config)
	api.SetConfig(&config)

	http.Handle("/", api.Commands())

	go func() {
		http.ListenAndServe(listen, nil)
	}()
	logrus.Fatal(mesos.Subscribe())
}
