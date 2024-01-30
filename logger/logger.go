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
	callPluginEvent("DEBUG", args, e.Value)
	logrus.WithField(e.Key, e.Value).Debug(args...)
}

func (e *Logger) Info(args ...interface{}) {
	callPluginEvent("INFO", args, e.Value)
	logrus.WithField(e.Key, e.Value).Info(args...)
}

func (e *Logger) Error(args ...interface{}) {
	callPluginEvent("ERROR", args, e.Value)
	logrus.WithField(e.Key, e.Value).Error(args...)
}

func (e *Logger) Warn(args ...interface{}) {
	callPluginEvent("WARN", args, e.Value)
	logrus.WithField(e.Key, e.Value).Warn(args...)
}

func (e *Logger) Warning(args ...interface{}) {
	callPluginEvent("WARN", args, e.Value)
	logrus.WithField(e.Key, e.Value).Warning(args...)
}

func (e *Logger) Warningf(format string, args ...interface{}) {
	callPluginEvent("WARN", args, e.Value)
	logrus.WithField(e.Key, e.Value).Warningf(format, args...)
}

func (e *Logger) Trace(args ...interface{}) {
	callPluginEvent("TRACE", args, e.Value)
	logrus.WithField(e.Key, e.Value).Trace(args...)
}

func (e *Logger) Fatal(args ...interface{}) {
	callPluginEvent("FATAL", args, e.Value)
	logrus.WithField(e.Key, e.Value).Fatal(args...)
}

func Fatal(text ...interface{}) {
	callPluginEvent("FATAL", text...)
	logrus.Fatal(text...)
}

func Println(text ...interface{}) {
	callPluginEvent("PLAIN", text...)
	logrus.Println(text...)
}

func SetLogLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

func SetPlugins(plugin map[string]*plugin.Plugin) {
	Plugins = plugin
}

func callPluginEvent(info string, args ...interface{}) {
	if len(Plugins) > 0 {
		for _, p := range Plugins {
			symbol, err := p.Lookup("Logging")
			if err != nil {
				continue
			}

			loggingPluginFunc, ok := symbol.(func(string, ...interface{}))
			if !ok {
				continue
			}

			loggingPluginFunc(info, args...)
		}
	}
}
