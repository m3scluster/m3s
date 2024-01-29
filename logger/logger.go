package logger

import (
	logrus "github.com/sirupsen/logrus"
)

func WithField(key string, value interface{}) *logrus.Logger {
	return logrus.WithField(key, value).Logger
}

func Fatal(text ...interface{}) {
	logrus.Fatal(text)
}

func Println(text ...interface{}) {
	logrus.Println(text)
}

func SetLogLevel(level logrus.Level) {
	logrus.SetLevel(level)
}
