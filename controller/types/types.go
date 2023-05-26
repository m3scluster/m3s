package types

import "time"

// Config is a struct of the redis configuration
type Config struct {
	AppName          string
	LogLevel         string
	RedisServer      string
	RedisPassword    string
	RedisDB          int
	RedisPrefix      string
	KubernetesConfig string
	Heartbeat        time.Duration
	DefaultYAML      string
}
