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
	K3SToken                    string
	ETCDMax                     int
	DockerSock                  string
	DockerSHMSize               string
	BootstrapURL                string
	M3SBootstrapServerHostname  string
	M3SBootstrapServerPort      int
	K3SServerCPU                float64
	K3SServerMEM                float64
	K3SServerDISK               float64
	K3SAgentCPU                 float64
	K3SAgentMEM                 float64
	K3SAgentDISK                float64
	K3SDocker                   string
	ETCDCPU                     float64
	ETCDMEM                     float64
	ETCDDISK                    float64
	ETCDConstraint              string
	ETCDConstraintHostname      string
	M3SStatus                   M3SStatus
	MesosSandboxVar             string
	RedisServer                 string
	RedisClient                 *goredis.Client
	RedisCTX                    context.Context
	RedisPassword               string
	RedisDB                     int
	SkipSSL                     bool
	SSLKey                      string
	SSLCrt                      string
	Version                     M3SVersion
	Suppress                    bool
	DockerCNI                   string
}

// M3SStatus store the current TaskState of the M3s services
type M3SStatus struct {
	Server map[string]string
	Agent  map[string]string
	API    string
	Etcd   map[string]string
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

// ErrorMsg hold the structure of error messages
type ErrorMsg struct {
	Message  string
	Number   int
	Function string
}

// M3SVersion hold the version numbers off the whole m3s stack
type M3SVersion struct {
	ClientVersion    versionInfo `json:"clientVersion"`
	ServerVersion    versionInfo `json:"serverVersion"`
	M3SVersion       versionInfo `json:"m3sVersion"`
	BootstrapVersion versionInfo `json:"bootstrapVersion"`
}

type versionInfo struct {
	Major        string `json:"major"`
	Minor        string `json:"minor"`
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}
