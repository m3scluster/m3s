package types

import mesosproto "mesos-k3s/proto"

// Config is a struct of the framework configuration
type Config struct {
	FrameworkPort              string
	FrameworkBind              string
	FrameworkUser              string
	FrameworkName              string
	FrameworkRole              string
	FrameworkInfo              mesosproto.FrameworkInfo
	FrameworkInfoFile          string
	FrameworkInfoFilePath      string
	Principal                  string
	Username                   string
	Password                   string
	MesosMasterServer          string
	MesosSSL                   bool
	MesosStreamID              string
	MesosCNI                   string
	TaskID                     uint64
	SSL                        bool
	LogLevel                   string
	MinVersion                 string
	AppName                    string
	EnableSyslog               bool
	Hostname                   string
	Listen                     string
	CommandChan                chan Command `json:"-"`
	State                      map[string]State
	Domain                     string
	K3SServerURL               string
	K3SAgentCount              int
	K3SAgentMax                int
	K3SServerCount             int
	K3SServerMax               int
	K3SServerPort              int
	K3SCustomDomain            string
	K3SServerString            string
	K3SAgentString             string
	K3SAgent                   map[string]string
	Credentials                UserCredentials
	ImageK3S                   string
	ImageETCD                  string
	VolumeDriver               string
	VolumeK3SServer            string
	PrefixHostname             string
	PrefixTaskName             string
	K3SToken                   string
	ETCDCount                  int
	ETCDMax                    int
	DockerSock                 string
	BootstrapURL               string
	M3SBootstrapServerHostname string
	M3SBootstrapServerPort     int
	K3SCPU                     float64
	K3SMEM                     float64
	ETCDCPU                    float64
	ETCDMEM                    float64
}

// Command is a chan which include all the Information about the started tasks
type Command struct {
	ContainerImage     string                                            `json:"container_image,omitempty"`
	ContainerType      string                                            `json:"container_type,omitempty"`
	TaskName           string                                            `json:"task_name,omitempty"`
	Command            string                                            `json:"command,omitempty"`
	Hostname           string                                            `json:"hostname,omitempty"`
	Domain             string                                            `json:"domain,omitempty"`
	Privileged         bool                                              `json:"privileged,omitempty"`
	NetworkMode        string                                            `json:"network_mode,omitempty"`
	Volumes            []mesosproto.Volume                               `protobuf:"bytes,2,rep,name=volumes" json:"volumes,omitempty"`
	Shell              bool                                              `protobuf:"varint,6,opt,name=shell,def=1" json:"shell,omitempty"`
	Uris               []mesosproto.CommandInfo_URI                      `protobuf:"bytes,1,rep,name=uris" json:"uris,omitempty"`
	Environment        mesosproto.Environment                            `protobuf:"bytes,2,opt,name=environment" json:"environment,omitempty"`
	NetworkInfo        []mesosproto.NetworkInfo                          `protobuf:"bytes,2,opt,name=networkinfo" json:"networkinfo,omitempty"`
	DockerPortMappings []mesosproto.ContainerInfo_DockerInfo_PortMapping `protobuf:"bytes,3,rep,name=port_mappings,json=portMappings" json:"port_mappings,omitempty"`
	DockerParameter    []mesosproto.Parameter                            `protobuf:"bytes,5,rep,name=parameters" json:"parameters,omitempty"`
	Arguments          []string                                          `protobuf:"bytes,7,rep,name=arguments" json:"arguments,omitempty"`
	Discovery          mesosproto.DiscoveryInfo                          `protobuf:"bytes,12,opt,name=discovery" json:"discovery,omitempty"`
	Executor           mesosproto.ExecutorInfo
	InternalID         int
	TaskID             uint64
	IsK3SAgent         bool
	IsK3SServer        bool
	IsETCD             bool
	Memory             float64
	CPU                float64
	Agent              string
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
