package scheduler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	util "github.com/AVENTER-UG/util/util"
	corev1 "k8s.io/api/core/v1"
)

// StartK3SAgent is starting a agent container with the given IDs
func (e *Scheduler) StartK3SAgent(taskID string) {

	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", "") == e.Config.K3SAgentMax {
		return
	}

	if e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", "") > e.Config.K3SAgentMax {
		e.API.Scale(e.Config.K3SAgentMax, e.Redis.CountRedisKey(e.Framework.FrameworkName+":agent:*", ""), "agent")
		return
	}

	cmd := e.defaultCommand(taskID)

	cmd.ContainerImage = e.Config.ImageK3S
	cmd.DockerPortMappings = []*mesosproto.ContainerInfo_DockerInfo_PortMapping{
		{
			HostPort:      util.Uint32ToPointer(0),
			ContainerPort: util.Uint32ToPointer(80),
			Protocol:      util.StringToPointer("http"),
		},
		{
			HostPort:      util.Uint32ToPointer(0),
			ContainerPort: util.Uint32ToPointer(443),
			Protocol:      util.StringToPointer("https"),
		},
	}

	if e.Config.K3SAgentTCPPort > 0 {
		tmpTcpPort := []*mesosproto.ContainerInfo_DockerInfo_PortMapping{
			{
				HostPort:      util.Uint32ToPointer(0),
				ContainerPort: util.Uint32ToPointer(uint32(e.Config.K3SAgentTCPPort)),
				Protocol:      util.StringToPointer("tcp"),
			},
		}
		cmd.DockerPortMappings = append(cmd.DockerPortMappings, tmpTcpPort...)
	}

	cmd.Shell = false
	cmd.Privileged = true
	cmd.Memory = e.Config.K3SAgentMEM
	cmd.CPU = e.Config.K3SAgentCPU
	cmd.Disk = e.Config.K3SAgentDISK
	cmd.CPULimit = e.Config.K3SAgentCPULimit
	cmd.MemoryLimit = e.Config.K3SAgentMEMLimit
	cmd.TaskName = e.Framework.FrameworkName + ":agent"
	cmd.Hostname = e.Framework.FrameworkName + "agent" + e.Config.Domain
	cmd.Command = "/mnt/mesos/sandbox/bootstrap"
	cmd.Arguments = strings.Split(e.Config.K3SAgentString, " ")
	if e.Config.K3SDocker != "" {
		cmd.Arguments = append(cmd.Arguments, e.Config.K3SDocker)
	}
	cmd.Arguments = append(cmd.Arguments, "--with-node-id "+cmd.TaskID)
	cmd.Arguments = append(cmd.Arguments, "--kubelet-arg node-labels m3s.aventer.biz/taskid="+cmd.TaskID)
	cmd.DockerParameter = e.addDockerParameter(make([]*mesosproto.Parameter, 0), "cap-add", "NET_ADMIN")
	cmd.DockerParameter = e.addDockerParameter(make([]*mesosproto.Parameter, 0), "cap-add", "SYS_ADMIN")
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "shm-size", e.Config.K3SContainerDisk)
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "memory-swap", fmt.Sprintf("%.0fg", (e.Config.DockerMemorySwap+e.Config.K3SAgentMEM)/1024))
	cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "ulimit", "nofile="+e.Config.DockerUlimit)

	for key, value := range e.Config.K3SAgentCustomDockerParameters {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, key, value)
	}

	if e.Config.RestrictDiskAllocation {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "storage-opt", fmt.Sprintf("size=%smb", strconv.Itoa(int(e.Config.K3SAgentDISKLimit))))
	}

	if e.Config.CustomDockerRuntime != "" {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "runtime", e.Config.CustomDockerRuntime)
	}

	if e.Config.EnableRegistryMirror {
		cmd.Arguments = append(cmd.Arguments, "--embedded-registry")
	}

	cmd.Instances = e.Config.K3SAgentMax

	// if mesos cni is unset, then use docker cni
	if e.Framework.MesosCNI == "" {
		// net-alias is only supported onuser-defined networks
		if e.Config.DockerCNI != "bridge" {
			cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "net-alias", e.Framework.FrameworkName+"agent")
		}
	}

	cmd.Uris = []*mesosproto.CommandInfo_URI{
		{
			Value:      &e.Config.BootstrapURL,
			Extract:    func() *bool { x := false; return &x }(),
			Executable: func() *bool { x := true; return &x }(),
			Cache:      func() *bool { x := false; return &x }(),
			OutputFile: func() *string { x := "bootstrap"; return &x }(),
		},
	}

	if e.Config.CGroupV2 {
		cmd.DockerParameter = e.addDockerParameter(cmd.DockerParameter, "cgroupns", "private")
	}

	cmd.Discovery = &mesosproto.DiscoveryInfo{
		Visibility: mesosproto.DiscoveryInfo_EXTERNAL.Enum(),
		Name:       &cmd.TaskName,
		Ports: &mesosproto.Ports{
			Ports: e.getDiscoveryInfoPorts(&cmd),
		},
	}

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
			Value: util.StringToPointer("agent"),
		},
		{
			Name:  util.StringToPointer("K3S_TOKEN"),
			Value: &e.Config.K3SToken,
		},
		{
			Name:  util.StringToPointer("K3S_URL"),
			Value: &e.Config.K3SServerURL,
		},
		{
			Name:  util.StringToPointer("MESOS_SANDBOX_VAR"),
			Value: &e.Config.MesosSandboxVar,
		},
		{
			Name:  util.StringToPointer("REDIS_SERVER"),
			Value: &e.Config.RedisServer,
		},
		{
			Name:  util.StringToPointer("REDIS_PASSWORD"),
			Value: &e.Config.RedisPassword,
		},
		{
			Name:  util.StringToPointer("REDIS_DB"),
			Value: util.StringToPointer(strconv.Itoa(e.Config.RedisDB)),
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

	for key, value := range e.Config.K3SAgentNodeEnvironmentVariable {
		env := &mesosproto.Environment_Variable{
			Name:  &key,
			Value: &value,
		}
		cmd.Environment.Variables = append(cmd.Environment.Variables, env)
	}

	if e.Config.K3SAgentLabels != nil {
		cmd.Labels = e.Config.K3SAgentLabels
	}

	if e.Config.K3SAgentLabels != nil {
		cmd.Labels = e.Config.K3SAgentLabels
	}

	// store mesos task in DB
	logrus.WithField("func", "scheduler.StartK3SAgent").Info("Schedule K3S Agent")
	e.Redis.SaveTaskRedis(&cmd)
}

// Get the discoveryinfo ports of the compose file
func (e *Scheduler) getDiscoveryInfoPorts(cmd *cfg.Command) []*mesosproto.Port {
	var disport []*mesosproto.Port
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
			cmd.DockerPortMappings[i].Protocol = util.StringToPointer("tcp")
		}

		disport = append(disport, &tmpport)
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
