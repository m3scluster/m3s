package util

import (
	"log/syslog"

	prefixed "github.com/AVENTER-UG/go-logrus-formatter"
	"github.com/sirupsen/logrus"
	logrusSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

// SetLogging sets the loglevel and text formating
func SetLogging(logLevel string, enableSyslog bool, name string) {

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return
	}

	logrus.SetFormatter(&prefixed.TextFormatter{
		ForceColors: true,
	})
	logrus.SetLevel(level)

	if enableSyslog {
		hook, _ := logrusSyslog.NewSyslogHook("", "", syslog.LOG_DEBUG, name)
		logrus.AddHook(hook)
	}
}
