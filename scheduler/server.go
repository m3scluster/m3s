package scheduler

import (
	"fmt"
	"strconv"
	"time"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	corev1 "k8s.io/api/core/v1"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// StartK3SServer Start K3S with the given id
func (e *Scheduler) StartK3SServer(taskID string) {
	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":server:*", "") >= e.Config.K3SServerMax {
		return
	}

	cmd := e.defaultCommand(taskID)

	cmd.Shell = true
	cmd.Privileged = true
	cmd.ContainerImage = e.Config.ImageK3S
	cmd.Memory = e.Config.K3SServerMEM
	cmd.CPU = e.Config.K3SServerCPU
	cmd.Disk = e.Config.K3SServerDISK
	cmd.TaskName = e.Framework.FrameworkName + ":server"
	cmd.Hostname = e.Framework.FrameworkName + "server" + e.Config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + e.Config.K3SServerString + e.Config.K3SDocker + " --tls-san=" + e.Framework.FrameworkName + "server" + "  --node-label m3s.aventer.biz/taskid=" + cmd.TaskID
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "NET_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "SYS_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "shm-size", Value: e.Config.K3SContainerDisk})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "memory-swap", Value: fmt.Sprintf("%.0fg", (e.Config.DockerMemorySwap+e.Config.K3SServerMEM)/1024)})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "ulimit", Value: "nofile=" + e.Config.DockerUlimit})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "cpus", Value: strconv.FormatFloat(e.Config.K3SServerCPU, 'f', -1, 64)})

	cmd.Instances = e.Config.K3SServerMax
	// if mesos cni is unset, then use docker cni
	if e.Framework.MesosCNI == "" {
		// net-alias is only supported onuser-defined networks
		if e.Config.DockerCNI != "bridge" {
			cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "net-alias", Value: e.Framework.FrameworkName + "server"})
		}
	}

	cmd.Uris = []mesosproto.CommandInfo_URI{
		{
			Value:      e.Config.BootstrapURL,
			Extract:    func() *bool { x := false; return &x }(),
			Executable: func() *bool { x := true; return &x }(),
			Cache:      func() *bool { x := false; return &x }(),
			OutputFile: func() *string { x := "bootstrap"; return &x }(),
		},
	}

	cmd.Volumes = []mesosproto.Volume{
		{
			ContainerPath: "/var/lib/rancher/k3s",
			Mode:          mesosproto.RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME,
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Driver: &e.Config.VolumeDriver,
					Name:   e.Config.VolumeK3SServer,
				},
			},
		},
	}

	if e.Config.CGroupV2 {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "cgroupns", Value: "host"})

		tmpVol := mesosproto.Volume{
			ContainerPath: "/sys/fs/cgroup",
			Mode:          mesosproto.RW.Enum(),
			Source: &mesosproto.Volume_Source{
				Type: mesosproto.Volume_Source_DOCKER_VOLUME,
				DockerVolume: &mesosproto.Volume_Source_DockerVolume{
					Driver: &e.Config.VolumeDriver,
					Name:   func() string { x := "/sys/fs/cgroup"; return x }(),
				},
			},
		}

		cmd.Volumes = append(cmd.Volumes, tmpVol)
	}

	protocol := "tcp"
	cmd.DockerPortMappings = []mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      uint32(e.Config.K3SServerPort),
			ContainerPort: 6443,
			Protocol:      &protocol,
		},
		{
			HostPort:      0,
			ContainerPort: 8080,
			Protocol:      &protocol,
		},
	}

	cmd.Discovery = mesosproto.DiscoveryInfo{
		Visibility: 2,
		Name:       &cmd.TaskName,
		Ports: &mesosproto.Ports{
			Ports: []mesosproto.Port{
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

	e.CreateK3SServerString()

	cmd.Environment.Variables = []mesosproto.Environment_Variable{
		{
			Name:  "SERVICE_NAME",
			Value: &cmd.TaskName,
		},
		{
			Name:  "KUBERNETES_VERSION",
			Value: &e.Config.KubernetesVersion,
		},
		{
			Name:  "K3SFRAMEWORK_TYPE",
			Value: func() *string { x := "server"; return &x }(),
		},
		{
			Name:  "K3S_URL",
			Value: &e.Config.K3SServerURL,
		},
		{
			Name:  "K3S_TOKEN",
			Value: &e.Config.K3SToken,
		},
		{
			Name:  "K3S_KUBECONFIG_MODE",
			Value: func() *string { x := "666"; return &x }(),
		},
		{
			Name:  "KUBECONFIG",
			Value: func() *string { x := e.Config.KubeConfig; return &x }(),
		},
		{
			Name:  "M3S_CONTROLLER__REDIS_SERVER",
			Value: func() *string { x := e.Config.RedisServer; return &x }(),
		},
		{
			Name:  "M3S_CONTROLLER__REDIS_PASSWORD",
			Value: func() *string { x := e.Config.RedisPassword; return &x }(),
		},
		{
			Name:  "M3S_CONTROLLER__REDIS_DB",
			Value: func() *string { x := strconv.Itoa(e.Config.RedisDB); return &x }(),
		},
		{
			Name:  "M3S_CONTROLLER__REDIS_PREFIX",
			Value: func() *string { x := e.Framework.FrameworkName; return &x }(),
		},
		{
			Name:  "M3S_CONTROLLER__LOGLEVEL",
			Value: func() *string { x := e.Config.LogLevel; return &x }(),
		},
		{
			Name: "M3S_CONTROLLER__ENABLE_TAINT",
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
			Name:  "MESOS_SANDBOX_VAR",
			Value: &e.Config.MesosSandboxVar,
		},
		{
			Name:  "TZ",
			Value: &e.Config.TimeZone,
		},
		{
			Name:  "MESOS_TASK_ID",
			Value: &cmd.TaskID,
		},
	}

	if e.Config.K3SServerLabels != nil {
		cmd.Labels = e.Config.K3SServerLabels
	}

	if e.Config.K3SServerLabels != nil {
		cmd.Labels = e.Config.K3SServerLabels
	}

	if e.Config.DSEtcd {
		ds := mesosproto.Environment_Variable{
			Name: "K3S_DATASTORE_ENDPOINT",
			Value: func() *string {
				x := "http://" + e.Framework.FrameworkName + "datastore" + e.Config.Domain + ":" + e.Config.DSPort
				return &x
			}(),
		}
		cmd.Environment.Variables = append(cmd.Environment.Variables, ds)
	}

	if e.Config.DSMySQL {
		ds := mesosproto.Environment_Variable{
			Name: "K3S_DATASTORE_ENDPOINT",
			Value: func() *string {
				x := "mysql://" + e.Config.DSMySQLUsername + ":" + e.Config.DSMySQLPassword + "@tcp(" + e.Framework.FrameworkName + "datastore" + e.Config.Domain + ":" + e.Config.DSPort + ")/k3s"
				return &x
			}(),
		}
		cmd.Environment.Variables = append(cmd.Environment.Variables, ds)

		// Enable TLS
		if e.Config.DSMySQLSSL {
			ds = mesosproto.Environment_Variable{
				Name: "K3S_DATASTORE_CAFILE",
				Value: func() *string {
					x := "/var/lib/rancher/k3s/ca.pem"
					return &x
				}(),
			}
			cmd.Environment.Variables = append(cmd.Environment.Variables, ds)

			ds = mesosproto.Environment_Variable{
				Name: "K3S_DATASTORE_CERTFILE",
				Value: func() *string {
					x := "/var/lib/rancher/k3s/client-cert.pem"
					return &x
				}(),
			}
			cmd.Environment.Variables = append(cmd.Environment.Variables, ds)

			ds = mesosproto.Environment_Variable{
				Name: "K3S_DATASTORE_KEYFILE",
				Value: func() *string {
					x := "/var/lib/rancher/k3s/client-key.pem"
					return &x
				}(),
			}
			cmd.Environment.Variables = append(cmd.Environment.Variables, ds)
		}
	}

	// store mesos task in DB
	logrus.WithField("func", "StartK3SServer").Info("Schedule K3S Server")
	e.Redis.SaveTaskRedis(cmd)
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

	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":"+kind+":*", "") < max {
		logrus.WithField("func", "scheduler.healthCheckNode").Warningf("K3s %s missing", kind)
		return false
	}

	// check the realstate of the kubernetes agents
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":" + kind + ":*")
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		task := e.Mesos.DecodeTask(key)
		node := e.Kubernetes.GetK8NodeFromTask(task)

		timeDiff := time.Since(task.StateTime).Minutes()
		if node.Name != "" {
			for _, status := range node.Status.Conditions {
				if status.Type == corev1.NodeReady {
					if status.Status == corev1.ConditionTrue {
						aState = true
					}
					if (status.Status == corev1.ConditionFalse || status.Status == corev1.ConditionUnknown) && timeDiff >= e.Config.K3SNodeTimeout.Minutes() {
						logrus.WithField("func", "scheduler.healthCheckNode").Warningf("K3S %s (%s) not ready", node.Name, task.TaskID)
						e.cleanupUnreadyTask(task)
						return false
					}
				}
			}
		} else {
			if timeDiff >= e.Config.K3SNodeTimeout.Minutes() {
				logrus.WithField("func", "scheduler.healthCheckNode").Warningf("K3S %s (%s) not ready", node.Name, task.TaskID)
				e.cleanupUnreadyTask(task)
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
