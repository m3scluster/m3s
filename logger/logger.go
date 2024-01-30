package logger

import (
	"plugin"

	logrus "github.com/sirupsen/logrus"
)

var (
	Plugins map[string]*plugin.Plugin
)

type Logger struct {
	Key   string
	Value interface{}
}

func WithField(key string, value interface{}) *Logger {
	e := &Logger{
		Key:   key,
		Value: value,
	}
	return e
}

func (e *Logger) Debug(args ...interface{}) {
	callPluginEvent(args...)
	logrus.WithField(e.Key, e.Value).Debug(args...)
}

func (e *Logger) Info(args ...interface{}) {
	callPluginEvent(args...)
	logrus.WithField(e.Key, e.Value).Info(args...)
}

func (e *Logger) Error(args ...interface{}) {
	logrus.WithField(e.Key, e.Value).Error(args...)
}

func (e *Logger) Warn(args ...interface{}) {
	logrus.WithField(e.Key, e.Value).Warn(args...)
}

func (e *Logger) Warning(args ...interface{}) {
	logrus.WithField(e.Key, e.Value).Warning(args...)
}

func (e *Logger) Warningf(format string, args ...interface{}) {
	logrus.WithField(e.Key, e.Value).Warningf(format, args...)
}

func (e *Logger) Trace(args ...interface{}) {
	logrus.WithField(e.Key, e.Value).Trace(args...)
}

func (e *Logger) Fatal(args ...interface{}) {
	logrus.WithField(e.Key, e.Value).Fatal(args...)
}

func Fatal(text ...interface{}) {
	logrus.Fatal(text...)
}

func Println(text ...interface{}) {
	logrus.Println(text...)
}

func SetLogLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

func SetPlugins(plugin map[string]*plugin.Plugin) {
	Plugins = plugin
}

func callPluginEvent(args ...interface{}) {
	if len(Plugins) > 0 {
		for _, p := range Plugins {
			symbol, err := p.Lookup("Logging")
			if err != nil {
				logrus.WithField("func", "logger.callPluginEvent").Error("Error lookup event function in plugin: ", err.Error())
				continue
			}

			loggingPluginFunc, ok := symbol.(func(...interface{}))
			if !ok {
				logrus.WithField("func", "main.callPluginEvent").Error("Error plugin does not have logging function")
				continue
			}

			loggingPluginFunc(args...)
		}
	}
}
