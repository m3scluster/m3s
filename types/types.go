package types

import (
	"context"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	goredis "github.com/go-redis/redis/v8"
)

// Config is a struct of the framework configuration
type Config struct {
	Principal                   string
	LogLevel                    string
	MinVersion                  string
	AppName                     string
	EnableSyslog                bool
	Hostname                    string
	Listen                      string
	Domain                      string
	K3SServerURL                string
	K3SAgentMax                 int
	K3SServerMax                int
	K3SServerPort               int
	K3SCustomDomain             string
	K3SServerString             string
	K3SAgentString              string
	K3SAgent                    map[string]string
	K3SAgentLabels              []mesosproto.Label
	K3SServerConstraint         string
	K3SServerConstraintHostname string
	K3SAgentConstraint          string
	K3SAgentConstraintHostname  string
	Credentials                 UserCredentials
	ImageK3S                    string
	ImageETCD                   string
	VolumeDriver                string
	VolumeK3SServer             string
	PrefixHostname              string
	K3SToken                    string
	ETCDMax                     int
	DockerSock                  string
	BootstrapURL                string
	M3SBootstrapServerHostname  string
	M3SBootstrapServerPort      int
	K3SCPU                      float64
	K3SMEM                      float64
	ETCDCPU                     float64
	ETCDMEM                     float64
	M3SStatus                   M3SStatus
	MesosSandboxVar             string
	RedisServer                 string
	RedisClient                 *goredis.Client
	RedisCTX                    context.Context
	RedisPassword               string
}

// M3SStatus store the current TaskState of the M3s services
type M3SStatus struct {
	Server []mesosproto.TaskState
	Agent  []mesosproto.TaskState
	API    string
	Etcd   []mesosproto.TaskState
}

// State will have the state of all tasks stated by this framework
type State struct {
	Command mesosutil.Command      `json:"command"`
	Status  *mesosproto.TaskStatus `json:"status"`
}

// UserCredentials - The Username and Password to authenticate against this framework
type UserCredentials struct {
	Username string
	Password string
}

// Count shows the current scale state of agent/server
type Count struct {
	Scale   int // how many should run
	Running int // how many are running
}

// Version shows the current version of m3s
type Version struct {
	BootstrapBuild string `json:"bootstrap_build"`
	M3sBuild       string `json:"m3s_build"`
	M3sBersion     string `json:"m3s_version"`
}

// ErrorMsg hold the structure of error messages
type ErrorMsg struct {
	Message  string
	Number   int
	Function string
}
