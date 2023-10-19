package scheduler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	corev1 "k8s.io/api/core/v1"

	"github.com/sirupsen/logrus"
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

	cmd.Shell = true
	cmd.Privileged = true
	cmd.Memory = e.Config.K3SAgentMEM
	cmd.CPU = e.Config.K3SAgentCPU
	cmd.Disk = e.Config.K3SAgentDISK
	cmd.TaskName = e.Framework.FrameworkName + ":agent"
	cmd.Hostname = e.Framework.FrameworkName + "agent" + e.Config.Domain
	cmd.Command = "$MESOS_SANDBOX/bootstrap '" + e.Config.K3SAgentString + e.Config.K3SDocker + " --with-node-id " + cmd.TaskID + "'"
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
	// Hold the at all state of the agent service.
	aState := false

	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", "") < e.Config.K3SAgentMax {
		logrus.WithField("func", "scheduler.healthCheckAgent").Warning("K3s Agent missing")
		return false
	}

	// check the realstate of the kubernetes agents
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":agent:*")
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		task := e.Mesos.DecodeTask(key)
		node := e.getK8NodeFromTask(task)

		if node.Name != "" {
			timeDiff := time.Since(node.CreationTimestamp.Time).Minutes()
			for _, status := range node.Status.Conditions {
				if status.Type == corev1.NodeReady {
					if status.Status == corev1.ConditionTrue {
						aState = true
					}
					if (status.Status == corev1.ConditionFalse || status.Status == corev1.ConditionUnknown) && timeDiff >= e.Config.K3SNodeTimeout.Minutes() {
						logrus.WithField("func", "scheduler.healthCheckAgent").Warning("K3S Agent not ready: " + node.Name + "(" + task.TaskID + ")")
						e.cleanupUnreadyTask(task)
						return false
					}
				}
			}
		} else {
			logrus.WithField("func", "scheduler.healthCheckAgent").Warning("K3S Agent not ready: ", task.TaskID)
			e.cleanupUnreadyTask(task)
			return false
		}
	}

	return aState
}

// cleanupUnreadyTask if a Mesos task is still unready after CleanupLoopTime Minutes, then it have to be removed.
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

// getTaskFromK8Node will give out the mesos task matched to the K8 node
func (e *Scheduler) getTaskFromK8Node(node corev1.Node) cfg.Command {
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":agent:*")
	for keys.Next(e.Redis.CTX) {
		taskID := e.getTaskIDFromAnnotation(node.Annotations)
		if taskID != "" {
			key := e.Redis.GetRedisKey(e.Framework.FrameworkName + ":agent:" + taskID)
			if key != "" {
				task := e.Mesos.DecodeTask(key)
				return task
			}
		}
	}
	return cfg.Command{}
}

// getK8NodeFromTask will give out the K8 node from mesos task
func (e *Scheduler) getK8NodeFromTask(task cfg.Command) corev1.Node {
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":kubernetes:*agent*")
	for keys.Next(e.Redis.CTX) {
		key := e.Redis.GetRedisKey(keys.Val())
		var k8Node corev1.Node
		err := json.NewDecoder(strings.NewReader(key)).Decode(&k8Node)
		if err != nil {
			logrus.WithField("func", "scheduler.getK8NodeFromTask").Error("Could not decode kubernetes node: ", err.Error())
			continue
		}
		taskID := e.getTaskIDFromAnnotation(k8Node.Annotations)
		if taskID == task.TaskID {
			return k8Node
		}
	}

	return corev1.Node{}
}

// getTaskIDFromAnnotation will return the Mesos Task ID in the annotation string
func (e *Scheduler) getTaskIDFromAnnotation(annotations map[string]string) string {
	for i, annotation := range annotations {
		if i == "k3s.io/node-args" {
			var args []string
			err := json.Unmarshal([]byte(annotation), &args)
			if err != nil {
				logrus.WithField("func", "scheduler.getTaskIDFromAnnotation").Error("Could not decode kubernetes node annotation: ", err.Error())
				continue
			}
			return args[len(args)-1]
		}
	}
	return ""
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
		task := e.getTaskFromK8Node(node)
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
