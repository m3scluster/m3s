package types

import (
	"plugin"
	"time"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	corev1 "k8s.io/api/core/v1"
)

// Config is a struct of the framework configuration
type Config struct {
	AppName                          string
	BootstrapCredentials             UserCredentials
	BootstrapSSLKey                  string
	BootstrapSSLCrt                  string
	BootstrapURL                     string
	Credentials                      UserCredentials
	CleanupLoopTime                  time.Duration
	CGroupV2                         bool
	DSMax                            int
	DSCPU                            float64
	DSMEM                            float64
	DSDISK                           float64
	DSDISKLimit                      float64
	DSConstraint                     string
	DSConstraintHostname             string
	DSPort                           string
	DSEtcd                           bool
	DSMySQL                          bool
	DSMySQLUsername                  string
	DSMySQLPassword                  string
	DSMySQLSSL                       bool
	DockerSock                       string
	DockerSHMSize                    string
	DockerMemorySwap                 float64
	DockerUlimit                     string
	DockerRunning                    bool
	Domain                           string
	DockerCNI                        string
	EventLoopTime                    time.Duration
	EnableSyslog                     bool
	Hostname                         string
	ImageK3S                         string
	ImageETCD                        string
	ImageMySQL                       string
	KubeConfig                       string
	KubernetesVersion                string
	K3SNodeTimeout                   time.Duration
	K3SNodeEnvironmentVariable       map[string]string
	K3SCustomDomain                  string
	K3SContainerDisk                 string
	K3SServerURL                     string
	K3SServerNodeEnvironmentVariable map[string]string
	K3SServerMax                     int
	K3SServerPort                    int
	K3SServerString                  string
	K3SServerConstraint              string
	K3SServerConstraintHostname      string
	K3SServerHostname                string
	K3SServerContainerPort           int
	K3SServerCPU                     float64
	K3SServerMEM                     float64
	K3SServerDISK                    float64
	K3SServerDISKLimit               float64
	K3SServerLabels                  []*mesosproto.Label
	K3SServerCustomDockerParameters  map[string]string
	K3SAgent                         map[string]string
	K3SAgentMax                      int
	K3SAgentString                   string
	K3SAgentNodeEnvironmentVariable  map[string]string
	K3SAgentLabels                   []*mesosproto.Label
	K3SAgentConstraint               string
	K3SAgentConstraintHostname       string
	K3SAgentCPU                      float64
	K3SAgentMEM                      float64
	K3SAgentDISK                     float64
	K3SAgentDISKLimit                float64
	K3SAgentTCPPort                  int
	K3SAgentCustomDockerParameters   map[string]string
	K3SDocker                        string
	K3SToken                         string
	K3SEnableTaint                   bool
	Listen                           string
	LogLevel                         string
	M3SStatus                        M3SStatus
	MesosSandboxVar                  string
	Principal                        string
	PluginsEnable                    bool
	Plugins                          map[string]*plugin.Plugin
	ReconcileLoopTime                time.Duration
	RefuseOffers                     float64
	ReviveLoopTime                   time.Duration
	RedisServer                      string
	RedisPassword                    string
	RedisDB                          int
	SkipSSL                          bool
	SSLKey                           string
	SSLCrt                           string
	VolumeDriver                     string
	VolumeK3SServer                  string
	VolumeDS                         string
	Version                          M3SVersion
	TimeZone                         string
	DSMaxRestore                     int
	K3SServerMaxRestore              int
	K3SAgentMaxRestore               int
	K3SDisableScheduling             bool
	HostConstraintsList              []string
	K3SServerCPULimit                float64
	K3SAgentCPULimit                 float64
	DSCPULimit                       float64
	K3SServerMEMLimit                float64
	K3SAgentMEMLimit                 float64
	DSMEMLimit                       float64
	EnforceMesosTaskLimits           bool
	RestrictDiskAllocation           bool
	EnableRegistryMirror             bool
	CustomDockerRuntime              string
}

// M3SStatus store the current TaskState of the M3s services
type M3SStatus struct {
	Server    map[string]K3SNodeStatus
	Agent     map[string]K3SNodeStatus
	API       string
	Datastore map[string]string
}

// M3sNodeStatus represents the MESOS TASK STATUS and its corresponding K3S Name
type K3SNodeStatus struct {
	Status      string
	K3sNodeName string
}

// State will have the state of all tasks stated by this framework
type State struct {
	Command Command                `json:"command"`
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
	M3SVersion versionInfo
	K3SVersion []K3SVersion
}

type K3SVersion struct {
	NodeName string
	NodeInfo corev1.NodeSystemInfo
}

type versionInfo struct {
	GitVersion string `json:"gitVersion"`
	BuildDate  string `json:"buildDate"`
}

// ETCDHealth hold the health state of the etcd service
type ETCDHealth struct {
	Health string `json:"health"`
	Reason string `json:"reason"`
}

// FrameworkConfig
type FrameworkConfig struct {
	FrameworkHostname     string
	FrameworkPort         string
	FrameworkBind         string
	FrameworkUser         string
	FrameworkName         string
	FrameworkRole         string
	FrameworkInfo         mesosproto.FrameworkInfo
	FrameworkInfoFile     string
	FrameworkInfoFilePath string
	PortRangeFrom         int
	PortRangeTo           int
	CommandChan           chan Command `json:"-"`
	Username              string
	Password              string
	MesosMasterServer     string
	MesosSSL              bool
	MesosStreamID         string
	MesosCNI              string
	TaskID                string
	SSL                   bool
	State                 map[string]State
}

// Command is a chan which include all the Information about the started tasks
type Command struct {
	ContainerImage     string                                             `json:"container_image,omitempty"`
	ContainerType      string                                             `json:"container_type,omitempty"`
	TaskName           string                                             `json:"task_name,omitempty"`
	Command            string                                             `json:"command,omitempty"`
	Hostname           string                                             `json:"hostname,omitempty"`
	Domain             string                                             `json:"domain,omitempty"`
	Privileged         bool                                               `json:"privileged,omitempty"`
	NetworkMode        string                                             `json:"network_mode,omitempty"`
	Volumes            []*mesosproto.Volume                               `protobuf:"bytes,1,rep,name=volumes" json:"volumes,omitempty"`
	Shell              bool                                               `protobuf:"varint,2,opt,name=shell,def=1" json:"shell,omitempty"`
	Uris               []*mesosproto.CommandInfo_URI                      `protobuf:"bytes,3,rep,name=uris" json:"uris,omitempty"`
	Environment        *mesosproto.Environment                            `protobuf:"bytes,4,opt,name=environment" json:"environment,omitempty"`
	NetworkInfo        []*mesosproto.NetworkInfo                          `protobuf:"bytes,5,opt,name=networkinfo" json:"networkinfo,omitempty"`
	DockerPortMappings []*mesosproto.ContainerInfo_DockerInfo_PortMapping `protobuf:"bytes,6,rep,name=port_mappings,json=portMappings" json:"port_mappings,omitempty"`
	DockerParameter    []*mesosproto.Parameter                            `protobuf:"bytes,7,rep,name=parameters" json:"parameters,omitempty"`
	Arguments          []string                                           `protobuf:"bytes,8,rep,name=arguments" json:"arguments,omitempty"`
	Discovery          *mesosproto.DiscoveryInfo                          `protobuf:"bytes,9,opt,name=discovery" json:"discovery,omitempty"`
	Executor           *mesosproto.ExecutorInfo                           `protobuf:"bytes,10,opt,name=executor" json:"executor,omitempty"`
	Restart            string
	TaskID             string
	Memory             float64
	MemoryLimit        float64
	CPU                float64
	CPULimit           float64
	Disk               float64
	Agent              string
	Labels             []*mesosproto.Label
	State              string
	StateTime          time.Time
	Instances          int
	LinuxInfo          *mesosproto.LinuxInfo `protobuf:"bytes,11,opt,name=linux_info,json=linuxInfo" json:"linux_info,omitempty"`
	PullPolicy         string
	MesosAgent         MesosSlaves
	EnableHealthCheck  bool
	Health             *mesosproto.HealthCheck
}

// MesosAgent
type MesosAgent struct {
	Slaves          []MesosSlaves `json:"slaves"`
	RecoveredSlaves []interface{} `json:"recovered_slaves"`
}

// MesosSlaves ..
type MesosSlaves struct {
	ID         string `json:"id"`
	Hostname   string `json:"hostname"`
	Port       int    `json:"port"`
	Attributes struct {
	} `json:"attributes"`
	Pid              string  `json:"pid"`
	RegisteredTime   float64 `json:"registered_time"`
	ReregisteredTime float64 `json:"reregistered_time"`
	Resources        struct {
		Disk  float64 `json:"disk"`
		Mem   float64 `json:"mem"`
		Gpus  float64 `json:"gpus"`
		Cpus  float64 `json:"cpus"`
		Ports string  `json:"ports"`
	} `json:"resources"`
	UsedResources struct {
		Disk  float64 `json:"disk"`
		Mem   float64 `json:"mem"`
		Gpus  float64 `json:"gpus"`
		Cpus  float64 `json:"cpus"`
		Ports string  `json:"ports"`
	} `json:"used_resources"`
	OfferedResources struct {
		Disk float64 `json:"disk"`
		Mem  float64 `json:"mem"`
		Gpus float64 `json:"gpus"`
		Cpus float64 `json:"cpus"`
	} `json:"offered_resources"`
	ReservedResources struct {
	} `json:"reserved_resources"`
	UnreservedResources struct {
		Disk  float64 `json:"disk"`
		Mem   float64 `json:"mem"`
		Gpus  float64 `json:"gpus"`
		Cpus  float64 `json:"cpus"`
		Ports string  `json:"ports"`
	} `json:"unreserved_resources"`
	Active                bool     `json:"active"`
	Deactivated           bool     `json:"deactivated"`
	Version               string   `json:"version"`
	Capabilities          []string `json:"capabilities"`
	ReservedResourcesFull struct {
	} `json:"reserved_resources_full"`
	UnreservedResourcesFull []struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		Scalar struct {
			Value float64 `json:"value"`
		} `json:"scalar,omitempty"`
		Role   string `json:"role"`
		Ranges struct {
			Range []struct {
				Begin int `json:"begin"`
				End   int `json:"end"`
			} `json:"range"`
		} `json:"ranges,omitempty"`
	} `json:"unreserved_resources_full"`
	UsedResourcesFull []struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		Scalar struct {
			Value float64 `json:"value"`
		} `json:"scalar,omitempty"`
		Role           string `json:"role"`
		AllocationInfo struct {
			Role string `json:"role"`
		} `json:"allocation_info"`
		Ranges struct {
			Range []struct {
				Begin int `json:"begin"`
				End   int `json:"end"`
			} `json:"range"`
		} `json:"ranges,omitempty"`
	} `json:"used_resources_full"`
	OfferedResourcesFull []interface{} `json:"offered_resources_full"`
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
				NetworkInfos []*mesosproto.NetworkInfo `json:"network_infos"`
			} `json:"container_status,omitempty"`
		} `json:"statuses"`
		Discovery mesosproto.DiscoveryInfo `json:"discovery"`
		Container mesosproto.ContainerInfo `json:"container"`
	} `json:"tasks"`
}

type K8Cluster struct {
	Server                string `yaml:"server"`
	InsecureSkipTLSVerify bool   `yaml:"insecure-skip-tls-verify,omitempty"`
}
type K8Clusters struct {
	Cluster K8Cluster `yaml:"cluster"`
	Name    string    `yaml:"name"`
}

type K8Context struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type K8Contexts struct {
	Context K8Context `yaml:"context"`
	Name    string    `yaml:"name"`
}

type K8User struct {
	Token string `yaml:"token"`
}
type K8Users struct {
	User K8User `yaml:"user"`
	Name string `yaml:"name"`
}

type K8KubeConfig struct {
	APIVersion     string       `yaml:"apiVersion"`
	Kind           string       `yaml:"kind"`
	Preferences    struct{}     `yaml:"preferences"`
	Clusters       []K8Clusters `yaml:"clusters"`
	Contexts       []K8Contexts `yaml:"contexts"`
	CurrentContext string       `yaml:"current-context"`
	Users          []K8Users    `yaml:"users"`
}
