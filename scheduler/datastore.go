package scheduler

import (
	"fmt"
	"strconv"
	"strings"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	util "github.com/AVENTER-UG/util/util"
)

// StartDatastore is starting the datastore container
func (e *Scheduler) StartDatastore(taskID string) {
	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":datastore:*", "") >= e.Config.DSMax {
		return
	}

	cmd := e.defaultCommand(taskID)

	cmd.ContainerType = "DOCKER"
	cmd.Privileged = false
	cmd.Memory = e.Config.DSMEM
	cmd.CPU = e.Config.DSCPU
	cmd.Disk = e.Config.DSDISK
	cmd.CPULimit = e.Config.DSCPULimit
	cmd.MemoryLimit = e.Config.DSMEMLimit
	cmd.TaskName = e.Framework.FrameworkName + ":datastore"
	cmd.Hostname = e.Framework.FrameworkName + "datastore" + e.Config.Domain
	cmd.DockerParameter = e.addDockerParameter(make([]*mesosproto.Parameter, 0), "cap-add", "NET_ADMIN")
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "memory-swap", fmt.Sprintf("%.0fg", (e.Config.DockerMemorySwap+e.Config.DSMEMLimit)/1024))
	cmd.Instances = e.Config.DSMax
	cmd.Shell = false

	// if mesos cni is unset, then use docker cni
	if e.Framework.MesosCNI == "" {
		// net-alias is only supported onuser-defined networks
		if e.Config.DockerCNI != "bridge" {
			cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "net-alias", "datastore")
		}
	}

	if e.Config.RestrictDiskAllocation {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "storage-opt", fmt.Sprintf("size=%smb", strconv.Itoa(int(e.Config.DSDISKLimit))))
	}

	if e.Config.CustomDockerRuntime != "" {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "runtime", e.Config.CustomDockerRuntime)
	}

	// if we use etcd as datastore
	if e.Config.DSEtcd {
		e.setETCD(&cmd)
	}

	// if we use mysql/maraidb as datastore
	if e.Config.DSMySQL {
		e.setMySQL(&cmd)
	}

	protocol := "tcp"
	containerPort, _ := strconv.ParseUint(e.Config.DSPort, 10, 32)
	cmd.DockerPortMappings = []*mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      util.Uint32ToPointer(0),
			ContainerPort: util.Uint32ToPointer(uint32(containerPort)),
			Protocol:      &protocol,
		},
	}

	cmd.Discovery = &mesosproto.DiscoveryInfo{
		Visibility: mesosproto.DiscoveryInfo_EXTERNAL.Enum(),
		Name:       &cmd.TaskName,
		Ports: &mesosproto.Ports{
			Ports: []*mesosproto.Port{
				{
					Number:   cmd.DockerPortMappings[0].HostPort,
					Name:     func() *string { x := "datastore"; return &x }(),
					Protocol: cmd.DockerPortMappings[0].Protocol,
				},
			},
		},
	}

	// store mesos task in DB
	logrus.WithField("func", "StartDatastore").Info("Schedule Datastore")
	e.Redis.SaveTaskRedis(&cmd)
}

// healthCheckDatastore check the health of all datastore ervers. Return true if all are fine.
func (e *Scheduler) healthCheckDatastore() bool {
	// Hold the at all state of the datastore service.
	dsState := false

	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":datastore:*")
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		task := e.Mesos.DecodeTask(key)
		if task.State == "TASK_RUNNING" && len(task.NetworkInfo) > 0 {
			// if the framework is running as container, and the task hostname is the same like the frameworks one,
			// then use the containerport instead of the random hostport
			dsState = true
		}
	}

	return dsState
}

// set mysql parameter of the mesos task
func (e *Scheduler) setMySQL(cmd *cfg.Command) {
	cmd.ContainerImage = e.Config.ImageMySQL
	//cmd.Shell = true
	// Enable TLS for Mariadb
	if e.Config.DSMySQLSSL {
		cmd.Arguments = e.appendString(make([]string, 0), "--ssl-ca=/var/lib/mysql/ca.pem")
		cmd.Arguments = e.appendString(cmd.Arguments, "--ssl-cert=/var/lib/mysql/server-cert.pem")
		cmd.Arguments = e.appendString(cmd.Arguments, "--ssl-key=/var/lib/mysql/server-key.pem")
	}
	cmd.Environment = &mesosproto.Environment{}
	cmd.Environment.Variables = []*mesosproto.Environment_Variable{
		{
			Name:  util.StringToPointer("SERVICE_NAME"),
			Value: &cmd.TaskName,
		},
		{
			Name:  util.StringToPointer("MYSQL_ROOT_PASSWORD"),
			Value: util.StringToPointer(e.Config.DSMySQLPassword),
		},
		{
			Name:  util.StringToPointer("MYSQL_DATABASE"),
			Value: util.StringToPointer("k3s"),
		},
		{
			Name:  util.StringToPointer("TZ"),
			Value: &e.Config.TimeZone,
		},
		{
			Name:  util.StringToPointer("MESOS_TASK_ID"),
			Value: &cmd.TaskID,
		},
	}
	cmd.Volumes = []*mesosproto.Volume{
		{
			ContainerPath: util.StringToPointer("/var/lib/mysql"),
			Mode:          mesosproto.Volume_RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME.Enum(),
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Driver: &e.Config.VolumeDriver,
					Name:   &e.Config.VolumeDS,
				},
			},
		},
	}

	cmd.EnableHealthCheck = true
	cmd.Health = &mesosproto.HealthCheck{}

	cmd.Health.Command = &mesosproto.CommandInfo{
		Shell:     util.BoolToPointer(false),
		Value:     util.StringToPointer("mysqladmin"),
		Arguments: strings.Split("ping -h localhost", " "),
	}
}

// set etcd parameter of the mesos task
func (e *Scheduler) setETCD(cmd *cfg.Command) {
	cmd.ContainerImage = e.Config.ImageETCD
	cmd.Command = "/usr/local/bin/etcd --listen-client-urls http://0.0.0.0:" + e.Config.DSPort + " --election-timeout '50000' --heartbeat-interval '5000'"
	cmd.Shell = true
	AdvertiseURL := "http://" + cmd.Hostname + ":" + e.Config.DSPort

	AllowNoneAuthentication := "yes"

	cmd.Environment = &mesosproto.Environment{}
	cmd.Environment.Variables = []*mesosproto.Environment_Variable{
		{
			Name:  util.StringToPointer("SERVICE_NAME"),
			Value: &cmd.TaskName,
		},
		{
			Name:  util.StringToPointer("ALLOW_NONE_AUTHENTICATION"),
			Value: &AllowNoneAuthentication,
		},
		{
			Name:  util.StringToPointer("ETCD_ADVERTISE_CLIENT_URLS"),
			Value: &AdvertiseURL,
		},
		{
			Name:  util.StringToPointer("MESOS_TASK_ID"),
			Value: &cmd.TaskID,
		},
		{
			Name:  util.StringToPointer("TZ"),
			Value: &e.Config.TimeZone,
		},
	}
	cmd.Volumes = []*mesosproto.Volume{
		{
			ContainerPath: util.StringToPointer("/default.etcd"),
			Mode:          mesosproto.Volume_RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME.Enum(),
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Driver: &e.Config.VolumeDriver,
					Name:   &e.Config.VolumeDS,
				},
			},
		},
	}

	cmd.EnableHealthCheck = true
	cmd.Health = &mesosproto.HealthCheck{}
	cmd.Health.DelaySeconds = func() *float64 { x := 60.0; return &x }()

	cmd.Health.Command = &mesosproto.CommandInfo{
		Shell:       util.BoolToPointer(true),
		Environment: cmd.Environment,
		Value:       util.StringToPointer("etcdctl endpoint health --endpoints=http://127.0.0.1:" + e.Config.DSPort),
	}
}
