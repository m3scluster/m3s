package types

import (
	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
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
	ImageMySQL                  string
	VolumeDriver                string
	VolumeK3SServer             string
	K3SToken                    string
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
	DSMax                       int
	DSCPU                       float64
	DSMEM                       float64
	DSDISK                      float64
	DSConstraint                string
	DSConstraintHostname        string
	DSPort                      string
	DSEtcd                      bool
	DSMySQL                     bool
	DSMySQLUsername             string
	DSMySQLPassword             string
	M3SStatus                   M3SStatus
	MesosSandboxVar             string
	RedisServer                 string
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

// ETCDHealth hold the health state of the etcd service
type ETCDHealth struct {
	Health string `json:"health"`
	Reason string `json:"reason"`
}

// MesosTasks hold the information of the task
type MesosTasks struct {
	Tasks []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		FrameworkID string `json:"framework_id"`
		ExecutorID  string `json:"executor_id"`
		SlaveID     string `json:"slave_id"`
		AgentID     string `json:"agent_id"`
		State       string `json:"state"`
		Resources   struct {
			Disk float64 `json:"disk"`
			Mem  float64 `json:"mem"`
			Gpus float64 `json:"gpus"`
			Cpus float64 `json:"cpus"`
		} `json:"resources"`
		Role     string `json:"role"`
		Statuses []struct {
			State           string  `json:"state"`
			Timestamp       float64 `json:"timestamp"`
			ContainerStatus struct {
				ContainerID struct {
					Value string `json:"value"`
				} `json:"container_id"`
				NetworkInfos []mesosproto.NetworkInfo `json:"network_infos"`
			} `json:"container_status,omitempty"`
		} `json:"statuses"`
		Discovery mesosproto.DiscoveryInfo `json:"discovery"`
		Container mesosproto.ContainerInfo `json:"container"`
	} `json:"tasks"`
}
