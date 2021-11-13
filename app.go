package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AVENTER-UG/mesos-m3s/api"
	"github.com/AVENTER-UG/mesos-m3s/mesos"
	mesosutil "github.com/AVENTER-UG/mesos-util"
	goredis "github.com/go-redis/redis/v8"

	util "github.com/AVENTER-UG/util"
	"github.com/sirupsen/logrus"
)

// MinVersion is the version number of this program
var MinVersion string

// init the redis cache
func initCache() {
	client := goredis.NewClient(&goredis.Options{
		Addr: config.RedisServer,
		DB:   0,
	})
	config.RedisCTX = context.Background()
	pong, err := client.Ping(config.RedisCTX).Result()
	logrus.Debug("Redis Health: ", pong, err)
	config.RedisClient = client
}

func main() {
	util.SetLogging(config.LogLevel, config.EnableSyslog, config.AppName)
	logrus.Println(config.AppName + " build " + MinVersion)

	listen := fmt.Sprintf(":%s", framework.FrameworkPort)

	failoverTimeout := 5000.0
	checkpoint := true
	webuiurl := fmt.Sprintf("http://%s%s", framework.FrameworkHostname, listen)

	framework.CommandChan = make(chan mesosutil.Command, 100)
	config.Hostname = framework.FrameworkHostname
	config.Listen = listen

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

	// The Hostname should ever be set after reading the state file.
	framework.FrameworkInfo.Hostname = &framework.FrameworkHostname

	initCache()

	mesosutil.SetConfig(&framework)
	mesos.SetConfig(&config, &framework)
	api.SetConfig(&config, &framework)

	// load framework state from DB
	key := api.GetRedisKey("framework")
	if key != "" {
		json.Unmarshal([]byte(key), &framework)
	}

	// load framework config from DB
	key = api.GetRedisKey("framework_config")
	if key != "" {
		json.Unmarshal([]byte(key), &config)
	}

	http.Handle("/", api.Commands())

	go func() {
		http.ListenAndServe(listen, nil)
	}()
	logrus.Fatal(mesos.Subscribe())
}
