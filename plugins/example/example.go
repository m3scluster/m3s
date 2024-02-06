package main

import (
	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	"github.com/AVENTER-UG/mesos-m3s/redis"
	"github.com/sirupsen/logrus"
)

type Plugins struct {
	Redis *redis.Redis
}

var plugin *Plugins

func Init(r *redis.Redis) string {
	plugin = &Plugins{
		Redis: r,
	}

	logrus.WithField("func", "plugin").Info("Example Plugin")

	return "example"
}

func Event(event mesosproto.Event) {
	logrus.WithField("func", "plugin").Debug("Example Plugin")
}
