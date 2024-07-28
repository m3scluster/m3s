package scheduler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	corev1 "k8s.io/api/core/v1"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// StartK3SAgent is starting a agent container with the given IDs
func (e *Scheduler) StartK3SAgent(taskID string) {
	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", "") >= e.Config.K3SAgentMax {
		return
	}

	cmd := e.defaultCommand(taskID)

	cmd.ContainerImage = e.Config.ImageK3S
	cmd.DockerPortMappings = []mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      0,
			ContainerPort: 80,
			Protocol:      func() *string { x := "http"; return &x }(),
		},
		{
			HostPort:      0,
			ContainerPort: 443,
			Protocol:      func() *string { x := "https"; return &x }(),
		},
	}

	if e.Config.K3SAgentTCPPort > 0 {
		tmpTcpPort := []mesosproto.ContainerInfo_DockerInfo_PortMapping{
			{
				HostPort:      0,
				ContainerPort: uint32(e.Config.K3SAgentTCPPort),
				Protocol:      func() *string { x := "tcp"; return &x }(),
			},
		}
		cmd.DockerPortMappings = append(cmd.DockerPortMappings, tmpTcpPort...)
	}

	cmd.Shell = true
	cmd.Privileged = true
	cmd.Memory = e.Config.K3SAgentMEM
	cmd.CPU = e.Config.K3SAgentCPU
	cmd.Disk = e.Config.K3SAgentDISK
	cmd.TaskName = e.Framework.FrameworkName + ":agent"
	cmd.Hostname = e.Framework.FrameworkName + "agent" + e.Config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + e.Config.K3SAgentString + e.Config.K3SDocker + " --with-node-id " + cmd.TaskID + " --node-label m3s.aventer.biz/taskid=" + cmd.TaskID + "'"
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "NET_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(make([]mesosproto.Parameter, 0), mesosproto.Parameter{Key: "cap-add", Value: "SYS_ADMIN"})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "shm-size", Value: e.Config.K3SContainerDisk})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "memory-swap", Value: fmt.Sprintf("%.0fg", (e.Config.DockerMemorySwap+e.Config.K3SAgentMEM)/1024)})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "ulimit", Value: "nofile=" + e.Config.DockerUlimit})
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "cpus", Value: strconv.FormatFloat(e.Config.K3SAgentCPU, 'f', -1, 64)})

	cmd.Instances = e.Config.K3SAgentMax

	// if mesos cni is unset, then use docker cni
	if e.Framework.MesosCNI == "" {
		// net-alias is only supported onuser-defined networks
		if e.Config.DockerCNI != "bridge" {
			cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "net-alias", Value: e.Framework.FrameworkName + "agent"})
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

	if e.Config.CGroupV2 {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, mesosproto.Parameter{Key: "cgroupns", Value: "host"})

		cmd.Volumes = []mesosproto.Volume{
			{
				ContainerPath: "/sys/fs/cgroup",
				Mode:          mesosproto.RW.Enum(),
				Source: &mesosproto.Volume_Source{
					Type: mesosproto.Volume_Source_DOCKER_VOLUME,
					DockerVolume: &mesosproto.Volume_Source_DockerVolume{
						Driver: &e.Config.VolumeDriver,
						Name:   func() string { x := "/sys/fs/cgroup"; return x }(),
					},
				},
			},
		}
	}

	cmd.Discovery = mesosproto.DiscoveryInfo{
		Visibility: 2,
		Name:       &cmd.TaskName,
		Ports: &mesosproto.Ports{
			Ports: e.getDiscoveryInfoPorts(cmd),
		},
	}

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
			Value: func() *string { x := "agent"; return &x }(),
		},
		{
			Name:  "K3S_TOKEN",
			Value: &e.Config.K3SToken,
		},
		{
			Name:  "K3S_URL",
			Value: &e.Config.K3SServerURL,
		},
		{
			Name:  "MESOS_SANDBOX_VAR",
			Value: &e.Config.MesosSandboxVar,
		},
		{
			Name:  "REDIS_SERVER",
			Value: &e.Config.RedisServer,
		},
		{
			Name:  "REDIS_PASSWORD",
			Value: &e.Config.RedisPassword,
		},
		{
			Name:  "REDIS_DB",
			Value: func() *string { x := strconv.Itoa(e.Config.RedisDB); return &x }(),
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

	if e.Config.K3SAgentLabels != nil {
		cmd.Labels = e.Config.K3SAgentLabels
	}

	if e.Config.K3SAgentLabels != nil {
		cmd.Labels = e.Config.K3SAgentLabels
	}

	// store mesos task in DB
	logrus.WithField("func", "scheduler.StartK3SAgent").Info("Schedule K3S Agent")
	e.Redis.SaveTaskRedis(cmd)
}

// Get the discoveryinfo ports of the compose file
func (e *Scheduler) getDiscoveryInfoPorts(cmd cfg.Command) []mesosproto.Port {
	var disport []mesosproto.Port
	for i, c := range cmd.DockerPortMappings {
		var tmpport mesosproto.Port
		p := func() *string {
			x := strings.ToLower(e.Framework.FrameworkName) + "-" + *c.Protocol
			return &x
		}()
		tmpport.Name = p
		tmpport.Number = c.HostPort
		tmpport.Protocol = c.Protocol

		// Docker understand only tcp and udp.
		if *c.Protocol != "udp" && *c.Protocol != "tcp" {
			cmd.DockerPortMappings[i].Protocol = func() *string { x := "tcp"; return &x }()
		}

		disport = append(disport, tmpport)
	}

	return disport
}

// healthCheckAgent check the health of all agents. Return true if all are fine.
func (e *Scheduler) healthCheckAgent() bool {
	return e.healthCheckNode("agent", e.Config.K3SAgentMax)
}

// removeNotExistingAgents remove kubernetes from redis if it does not have a Mesos Task. It
// will also kill the Mesos Task if the Agent is unready but the Task is in state RUNNING.
func (e *Scheduler) removeNotExistingAgents() {
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":kubernetes:*agent*")
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		var node corev1.Node
		err := json.NewDecoder(strings.NewReader(key)).Decode(&node)
		if err != nil {
			logrus.WithField("func", "scheduler.removeNotExistingAgents").Error("Could not decode kubernetes node: ", err.Error())
			continue
		}
		task := e.Kubernetes.GetTaskFromK8Node(node, "agent")
		if task.TaskID != "" {
			for _, status := range node.Status.Conditions {
				if status.Type == corev1.NodeReady && status.Status == corev1.ConditionUnknown && task.State == "TASK_RUNNING" {
					logrus.WithField("func", "scheduler.removeNotExistingAgents").Debug("Kill unready Agents: ", node.Name)
					e.Mesos.Kill(task.TaskID, task.Agent)
				}
			}
		} else {
			logrus.WithField("func", "scheduler.removeNotExistingAgents").Debug("Remove K8s Agent that does not have running Mesos task: ", node.Name)
			e.Redis.DelRedisKey(e.Framework.FrameworkName + ":kubernetes:" + node.Name)
		}
	}
}
