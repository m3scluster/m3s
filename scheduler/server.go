package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	"github.com/AVENTER-UG/util/util"
	corev1 "k8s.io/api/core/v1"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// StartK3SServer Start K3S with the given id
func (e *Scheduler) StartK3SServer(taskID string) {
	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":server:*", "") == e.Config.K3SServerMax {
		return
	}

	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":server:*", "") > e.Config.K3SServerMax {
		e.API.Scale(e.Config.K3SServerMax, e.Redis.CountRedisKey(e.Framework.FrameworkName+":server:*", ""), "server")
		return
	}

	cmd := e.defaultCommand(taskID)

	cmd.Shell = false
	cmd.Privileged = true
	cmd.ContainerImage = e.Config.ImageK3S
	cmd.Memory = e.Config.K3SServerMEM
	cmd.CPU = e.Config.K3SServerCPU
	cmd.CPULimit = e.Config.K3SServerCPULimit
	cmd.MemoryLimit = e.Config.K3SServerMEMLimit
	cmd.Disk = e.Config.K3SServerDISK
	cmd.TaskName = e.Framework.FrameworkName + ":server"
	cmd.Hostname = e.Framework.FrameworkName + "server" + e.Config.Domain
	cmd.Command = "/mnt/mesos/sandbox/bootstrap"
	cmd.Arguments = strings.Split(e.Config.K3SServerString, " ")
	if e.Config.K3SDocker != "" {
		cmd.Arguments = append(cmd.Arguments, e.Config.K3SDocker)
	}
	cmd.Arguments = append(cmd.Arguments, "--tls-san="+e.Framework.FrameworkName+"server")
	cmd.Arguments = append(cmd.Arguments, "--node-label m3s.aventer.biz/taskid="+cmd.TaskID)
	if e.Config.K3SEnableTaint {
		cmd.Arguments = append(cmd.Arguments, "--node-taint node-role.kubernetes.io/master=NoSchedule:NoSchedule")
	}
	cmd.DockerParameter = e.addDockerParameter(make([]*mesosproto.Parameter, 0), "cap-add", "NET_ADMIN")
	cmd.DockerParameter = e.addDockerParameter(make([]*mesosproto.Parameter, 0), "cap-add", "SYS_ADMIN")
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "shm-size", e.Config.K3SContainerDisk)
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "memory-swap", fmt.Sprintf("%.0fg", (e.Config.DockerMemorySwap+e.Config.K3SServerMEM)/1024))
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "ulimit", "nofile="+e.Config.DockerUlimit)

	for key, value := range e.Config.K3SServerCustomDockerParameters {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, key, value)
	}

	if e.Config.RestrictDiskAllocation {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "storage-opt", fmt.Sprintf("size=%smb", strconv.Itoa(int(e.Config.K3SServerDISKLimit))))
	}

	if e.Config.CustomDockerRuntime != "" {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "runtime", e.Config.CustomDockerRuntime)
	}

	cmd.Instances = e.Config.K3SServerMax
	// if mesos cni is unset, then use docker cni
	if e.Framework.MesosCNI == "" {
		// net-alias is only supported on user-defined networks
		if e.Config.DockerCNI != "bridge" {
			cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "net-alias", e.Framework.FrameworkName+"server")
		}
	}

	cmd.Uris = []*mesosproto.CommandInfo_URI{
		{
			Value:      &e.Config.BootstrapURL,
			Extract:    func() *bool { x := false; return &x }(),
			Executable: func() *bool { x := true; return &x }(),
			Cache:      func() *bool { x := false; return &x }(),
			OutputFile: util.StringToPointer("bootstrap"),
		},
	}

	cmd.Volumes = []*mesosproto.Volume{
		{
			ContainerPath: util.StringToPointer("/var/lib/rancher/k3s"),
			Mode:          mesosproto.Volume_RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME.Enum(),
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Driver: &e.Config.VolumeDriver,
					Name:   &e.Config.VolumeK3SServer,
				},
			},
		},
	}

	if e.Config.CGroupV2 {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "cgroupns", "private")
	}

	protocol := "tcp"
	cmd.DockerPortMappings = []*mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      util.Uint32ToPointer(uint32(e.Config.K3SServerPort)),
			ContainerPort: util.Uint32ToPointer(6443),
			Protocol:      &protocol,
		},
		{
			HostPort:      util.Uint32ToPointer(0),
			ContainerPort: util.Uint32ToPointer(8080),
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
					Name:     func() *string { x := "kubernetes"; return &x }(),
					Protocol: cmd.DockerPortMappings[0].Protocol,
				},
				{
					Number:   cmd.DockerPortMappings[1].HostPort,
					Name:     func() *string { x := "http"; return &x }(),
					Protocol: cmd.DockerPortMappings[1].Protocol,
				},
			},
		},
	}

	if e.Config.EnableRegistryMirror {
		cmd.Arguments = append(cmd.Arguments, "--embedded-registry")
	}

	e.CreateK3SServerString()

	cmd.Environment = &mesosproto.Environment{}
	cmd.Environment.Variables = []*mesosproto.Environment_Variable{
		{
			Name:  util.StringToPointer("SERVICE_NAME"),
			Value: &cmd.TaskName,
		},
		{
			Name:  util.StringToPointer("KUBERNETES_VERSION"),
			Value: &e.Config.KubernetesVersion,
		},
		{
			Name:  util.StringToPointer("K3SFRAMEWORK_TYPE"),
			Value: util.StringToPointer("server"),
		},
		{
			Name:  util.StringToPointer("K3S_URL"),
			Value: &e.Config.K3SServerURL,
		},
		{
			Name:  util.StringToPointer("K3S_TOKEN"),
			Value: &e.Config.K3SToken,
		},
		{
			Name:  util.StringToPointer("K3S_KUBECONFIG_MODE"),
			Value: func() *string { x := "666"; return &x }(),
		},
		{
			Name:  util.StringToPointer("KUBECONFIG"),
			Value: &e.Config.KubeConfig,
		},
		{
			Name:  util.StringToPointer("M3S_CONTROLLER__REDIS_SERVER"),
			Value: &e.Config.RedisServer,
		},
		{
			Name:  util.StringToPointer("M3S_CONTROLLER__REDIS_PASSWORD"),
			Value: &e.Config.RedisPassword,
		},
		{
			Name:  util.StringToPointer("M3S_CONTROLLER__REDIS_DB"),
			Value: util.StringToPointer(strconv.Itoa(e.Config.RedisDB)),
		},
		{
			Name:  util.StringToPointer("M3S_CONTROLLER__REDIS_PREFIX"),
			Value: &e.Framework.FrameworkName,
		},
		{
			Name:  util.StringToPointer("M3S_CONTROLLER__LOGLEVEL"),
			Value: &e.Config.LogLevel,
		},
		{
			Name: util.StringToPointer("M3S_CONTROLLER__ENABLE_TAINT"),
			Value: func() *string {
				x := "false"
				if e.Config.K3SEnableTaint {
					x = "true"
					return &x
				}
				return &x
			}(),
		},
		{
			Name:  util.StringToPointer("MESOS_SANDBOX_VAR"),
			Value: &e.Config.MesosSandboxVar,
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

	for key, value := range e.Config.K3SNodeEnvironmentVariable {
		env := &mesosproto.Environment_Variable{
			Name:  &key,
			Value: &value,
		}
		cmd.Environment.Variables = append(cmd.Environment.Variables, env)
	}

	for key, value := range e.Config.K3SServerNodeEnvironmentVariable {
		env := &mesosproto.Environment_Variable{
			Name:  &key,
			Value: &value,
		}
		cmd.Environment.Variables = append(cmd.Environment.Variables, env)
	}

	if e.Config.K3SServerLabels != nil {
		cmd.Labels = e.Config.K3SServerLabels
	}

	if e.Config.K3SServerLabels != nil {
		cmd.Labels = e.Config.K3SServerLabels
	}

	if e.Config.DSEtcd {
		ds := mesosproto.Environment_Variable{
			Name:  util.StringToPointer("K3S_DATASTORE_ENDPOINT"),
			Value: util.StringToPointer("http://" + e.Framework.FrameworkName + "datastore" + e.Config.Domain + ":" + e.Config.DSPort),
		}
		cmd.Environment.Variables = append(cmd.Environment.Variables, &ds)
	}

	if e.Config.DSMySQL {
		ds := mesosproto.Environment_Variable{
			Name:  util.StringToPointer("K3S_DATASTORE_ENDPOINT"),
			Value: util.StringToPointer("mysql://" + e.Config.DSMySQLUsername + ":" + e.Config.DSMySQLPassword + "@tcp(" + e.Framework.FrameworkName + "datastore" + e.Config.Domain + ":" + e.Config.DSPort + ")/k3s"),
		}
		cmd.Environment.Variables = append(cmd.Environment.Variables, &ds)

		// Enable TLS
		if e.Config.DSMySQLSSL {
			ds = mesosproto.Environment_Variable{
				Name:  util.StringToPointer("K3S_DATASTORE_CAFILE"),
				Value: util.StringToPointer("/var/lib/rancher/k3s/ca.pem"),
			}
			cmd.Environment.Variables = append(cmd.Environment.Variables, &ds)

			ds = mesosproto.Environment_Variable{
				Name:  util.StringToPointer("K3S_DATASTORE_CERTFILE"),
				Value: util.StringToPointer("/var/lib/rancher/k3s/client-cert.pem"),
			}
			cmd.Environment.Variables = append(cmd.Environment.Variables, &ds)

			ds = mesosproto.Environment_Variable{
				Name:  util.StringToPointer("K3S_DATASTORE_KEYFILE"),
				Value: util.StringToPointer("/var/lib/rancher/k3s/client-key.pem"),
			}
			cmd.Environment.Variables = append(cmd.Environment.Variables, &ds)
		}
	}

	// store mesos task in DB
	logrus.WithField("func", "StartK3SServer").Info("Schedule K3S Server")
	e.Redis.SaveTaskRedis(&cmd)
}

// CreateK3SServerString create the K3S_URL string
func (e *Scheduler) CreateK3SServerString() {
	server := "https://" + e.Framework.FrameworkName + "server" + e.Config.Domain + ":6443"

	e.Config.K3SServerURL = server
}

// healthCheckK3s check if the kubernetes server is already running
func (e *Scheduler) healthCheckK3s() bool {
	return e.healthCheckNode("server", e.Config.K3SServerMax)
}

// healthCheckNode check if the kubernetes node is already running and in ready state
func (e *Scheduler) healthCheckNode(kind string, max int) bool {
	// Hold the at all state of the agent service.
	aState := false

	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":"+kind+":*", "") != max {
		logrus.WithField("func", "scheduler.healthCheckNode").Warningf("K3s %s missing", kind)
		return false
	}

	// check the realstate of the kubernetes agents
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":" + kind + ":*")
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		task := e.Mesos.DecodeTask(key)
		node := e.Kubernetes.GetK8NodeFromTask(*task)

		timeDiff := time.Since(task.StateTime).Minutes()
		if node.Name != "" {
			for _, status := range node.Status.Conditions {
				if status.Type == corev1.NodeReady {
					if status.Status == corev1.ConditionTrue {
						aState = true
					}
					if (status.Status == corev1.ConditionFalse || status.Status == corev1.ConditionUnknown) && timeDiff >= e.Config.K3SNodeTimeout.Minutes() {
						logrus.WithField("func", "scheduler.healthCheckNode").Warningf("K3S %s (%s) not ready", node.Name, task.TaskID)
						e.cleanupUnreadyTask(*task)
						return false
					}
				}
			}
		} else {
			if timeDiff >= e.Config.K3SNodeTimeout.Minutes() {
				logrus.WithField("func", "scheduler.healthCheckNode").Warningf("K3S %s (%s) not ready", node.Name, task.TaskID)
				e.cleanupUnreadyTask(*task)
			}
			return false
		}
	}

	return aState
}

// cleanupUnreadyTask if a Mesos task is still unready after CleanupLoopTime Minutes, then it has to be removed.
func (e *Scheduler) cleanupUnreadyTask(task cfg.Command) {
	timeDiff := time.Since(task.StateTime).Minutes()
	if timeDiff >= e.Config.CleanupLoopTime.Minutes() {
		if task.MesosAgent.ID == "" {
			e.Redis.DelRedisKey(task.TaskName + ":" + task.TaskID)
		} else {
			e.Mesos.Kill(task.TaskID, task.MesosAgent.ID)
		}
		logrus.WithField("func", "scheduler.cleanupUnreadyTask").Warningf("Cleanup Unhealthy Mesos Task: %s", task.TaskID)
	}
}
