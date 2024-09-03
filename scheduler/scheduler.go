package scheduler

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net/http"
	"strconv"
	"strings"

	api "github.com/AVENTER-UG/mesos-m3s/api"
	"github.com/AVENTER-UG/mesos-m3s/controller"
	"github.com/AVENTER-UG/mesos-m3s/mesos"
	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	"github.com/AVENTER-UG/mesos-m3s/redis"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	"github.com/AVENTER-UG/util/util"
	"google.golang.org/protobuf/encoding/protojson"

	logrus "github.com/AVENTER-UG/mesos-m3s/logger"
)

// Scheduler include all the current vars and global config
type Scheduler struct {
	Config     *cfg.Config
	Framework  *cfg.FrameworkConfig
	Mesos      mesos.Mesos
	Client     *http.Client
	Req        *http.Request
	API        *api.API
	Redis      *redis.Redis
	Kubernetes *controller.Controller
}

var marshaller = protojson.MarshalOptions{
	UseEnumNumbers: false,
	Indent:         " ",
	UseProtoNames:  true,
}

// Subscribe to the mesos backend
func Subscribe(cfg *cfg.Config, frm *cfg.FrameworkConfig) *Scheduler {
	e := &Scheduler{
		Config:    cfg,
		Framework: frm,
		Mesos:     *mesos.New(cfg, frm),
	}

	subscribeCall := &mesosproto.Call{
		FrameworkId: e.Framework.FrameworkInfo.Id,
		Type:        mesosproto.Call_SUBSCRIBE.Enum(),
		Subscribe: &mesosproto.Call_Subscribe{
			FrameworkInfo: &e.Framework.FrameworkInfo,
		},
	}

	if e.Config.EnableHostnameOfferConstraint {

		offerConstraintGroups := []*mesosproto.OfferConstraints_RoleConstraints_Group{}
		for _, hostname := range e.Config.HostConstraintsList {
			offerConstraint := mesosproto.OfferConstraints_RoleConstraints_Group{
				AttributeConstraints: []*mesosproto.AttributeConstraint{
					{
						Selector: &mesosproto.AttributeConstraint_Selector{
							Selector: &mesosproto.AttributeConstraint_Selector_PseudoattributeType_{
								PseudoattributeType: mesosproto.AttributeConstraint_Selector_HOSTNAME,
							},
						},
						Predicate: &mesosproto.AttributeConstraint_Predicate{
							Predicate: &mesosproto.AttributeConstraint_Predicate_TextEquals_{
								TextEquals: &mesosproto.AttributeConstraint_Predicate_TextEquals{
									Value: &hostname,
								},
							},
						},
					},
				},
			}
			offerConstraintGroups = append(offerConstraintGroups, &offerConstraint)
		}

		offerConstraints := &mesosproto.OfferConstraints{
			RoleConstraints: map[string]*mesosproto.OfferConstraints_RoleConstraints{
				frm.FrameworkRole: {
					Groups: offerConstraintGroups,
				},
			},
		}

		subscribeCall.Subscribe.OfferConstraints = offerConstraints
	}

	logrus.WithField("func", "scheduler.Subscribe").Debug(subscribeCall)
	body, _ := marshaller.Marshal(subscribeCall)
	client := &http.Client{}
	// #nosec G402
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: e.Config.SkipSSL},
	}

	protocol := "https"
	if !e.Framework.MesosSSL {
		protocol = "http"
	}
	req, _ := http.NewRequest("POST", protocol+"://"+e.Framework.MesosMasterServer+"/api/v1/scheduler", bytes.NewBuffer([]byte(body)))
	req.Close = true
	req.SetBasicAuth(e.Framework.Username, e.Framework.Password)
	req.Header.Set("Content-Type", "application/json")

	e.Req = req
	e.Client = client

	return e
}

// EventLoop is the main loop for the mesos events.
func (e *Scheduler) EventLoop() {
	res, err := e.Client.Do(e.Req)

	if err != nil {
		logrus.WithField("func", "scheduler.EventLoop").Error("Mesos Master connection error: ", err.Error())
		return
	}
	defer res.Body.Close()

	reader := bufio.NewReader(res.Body)

	line, _ := reader.ReadString('\n')
	bytesCount, _ := strconv.Atoi(strings.Trim(line, "\n"))

	go e.HeartbeatLoop()
	go e.ReconcileLoop()

	for {
		// Read line from Mesos
		line, err = reader.ReadString('\n')
		_ = strings.Trim(line, "\n")
		if err != nil {
			logrus.WithField("func", "scheduler.EventLoop").Error("Error to read data from Mesos Master: ", err.Error())
			return
		}

		// skip if no data
		if line == "" || len(line)-1 < bytesCount {
			logrus.WithField("func", "scheduler.EventLoop").Tracef("Line %s, bytesCount: %d ", line, bytesCount)
			logrus.WithField("func", "scheduler.EventLoop").Trace("No data from Mesos Master")
			continue
		}
		data := line[:bytesCount]
		bytesCount, _ = strconv.Atoi(line[bytesCount : len(line)-1])

		// Read important data
		var event mesosproto.Event // Event as ProtoBuf
		err := protojson.Unmarshal([]byte(data), &event)
		if err != nil {
			logrus.WithField("func", "scheduler.EventLoop").Warn("Could not unmarshal Mesos Master data: ", err.Error())
			continue
		}

		logrus.WithField("func", "scheduler.EventLoop").Debug("Event Received from Mesos: ", event.Type)

		switch event.Type.Number() {
		case mesosproto.Event_SUBSCRIBED.Number():
			logrus.WithField("func", "scheduler.EventLoop").Info("Subscribed")
			logrus.WithField("func", "scheduler.EventLoop").Debug("FrameworkId: ", event.Subscribed.GetFrameworkId())
			e.Framework.FrameworkInfo.Id = event.Subscribed.GetFrameworkId()
			e.Framework.MesosStreamID = res.Header.Get("Mesos-Stream-Id")

			e.reconcile()
			e.CheckState()
			go e.callPluginEvent(&event)
			e.Redis.SaveFrameworkRedis(e.Framework)
			e.Redis.SaveConfig(*e.Config)
		case mesosproto.Event_UPDATE.Number():
			e.HandleUpdate(&event)
			// save configuration
			e.Redis.SaveConfig(*e.Config)
			go e.callPluginEvent(&event)
		case mesosproto.Event_OFFERS.Number():
			// Search Failed containers and restart them
			err = e.HandleOffers(event.Offers)
			if err != nil {
				logrus.WithField("func", "scheduler.EventLoop").Warn("Switch Event HandleOffers: ", err)
			}
		}
	}
}

// Generate "num" numbers of random host portnumbers
func (e *Scheduler) getRandomHostPort(num int) uint32 {
	// search two free ports
	for i := (e.Framework.PortRangeFrom + 1); i < e.Framework.PortRangeTo; i++ {
		port := uint32(i)
		use := false
		for x := 0; x < num; x++ {
			if e.portInUse(port + uint32(x)) {
				tmp := use || true
				use = tmp
				x = num
			}

			tmp := use || false
			use = tmp
		}
		if !use {
			return port
		}
	}
	return 0
}

// Check if the port is already in use
func (e *Scheduler) portInUse(port uint32) bool {
	// get all running services
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":*")
	for keys.Next(e.Redis.CTX) {
		// get the details of the current running service
		key := e.Redis.GetRedisKey(keys.Val())

		if e.Redis.CheckIfNotTask(keys) {
			continue
		}

		task := e.Mesos.DecodeTask(key)

		// check if the given port is already in use
		if task.Discovery != nil {
			ports := task.Discovery.Ports
			for _, hostport := range ports.GetPorts() {
				if hostport.GetNumber() == port {
					logrus.WithField("func", "scheduler.portInUse").Debug("Port in use: ", port)
					return true
				}
			}
		}
	}
	return false
}

// generate a new task id if there is no
func (e *Scheduler) getTaskID(taskID string) string {
	// if taskID is 0, then its a new task and we have to create a new ID
	if taskID == "" {
		taskID, _ = util.GenUUID()
	}

	return taskID
}

func (e *Scheduler) addDockerParameter(current []*mesosproto.Parameter, key string, values string) []*mesosproto.Parameter {
	newValues := &mesosproto.Parameter{
		Key:   func() *string { x := key; return &x }(),
		Value: func() *string { x := values; return &x }(),
	}

	return append(current, newValues)
}

func (e *Scheduler) appendString(current []string, newValues string) []string {
	return append(current, newValues)
}

func (e *Scheduler) changeDockerPorts(cmd *cfg.Command) []*mesosproto.ContainerInfo_DockerInfo_PortMapping {
	var ret []*mesosproto.ContainerInfo_DockerInfo_PortMapping
	hostPort := e.getRandomHostPort(len(cmd.Discovery.Ports.Ports))
	for n, port := range cmd.DockerPortMappings {
		if *port.HostPort == 0 {
			port.HostPort = util.Uint32ToPointer(hostPort + uint32(n))
		}
		ret = append(ret, port)
	}
	return ret
}

func (e *Scheduler) changeDiscoveryInfo(cmd *cfg.Command) *mesosproto.DiscoveryInfo {
	for i, port := range cmd.DockerPortMappings {
		cmd.Discovery.Ports.Ports[i].Number = port.HostPort
	}
	return cmd.Discovery
}

// reconcile will ask Mesos about the current state of the given tasks
func (e *Scheduler) reconcile() {
	var oldTasks []*mesosproto.Call_Reconcile_Task
	keys := e.Redis.GetAllRedisKeys(e.Framework.FrameworkName + ":*")
	for keys.Next(e.Redis.CTX) {
		// continue if the key is not a mesos task

		if e.Redis.CheckIfNotTask(keys) {
			continue
		}

		key := e.Redis.GetRedisKey(keys.Val())

		task := e.Mesos.DecodeTask(key)

		if task.TaskID == "" || task.Agent == "" {
			continue
		}

		oldTasks = append(oldTasks, &mesosproto.Call_Reconcile_Task{
			TaskId: &mesosproto.TaskID{
				Value: &task.TaskID,
			},
			AgentId: &mesosproto.AgentID{
				Value: &task.MesosAgent.ID,
			},
		})
		logrus.WithField("func", "mesos.Reconcile").Debug("Reconcile Task: ", task.TaskID, task.TaskName)
	}
	err := e.Mesos.Call(&mesosproto.Call{
		Type:      mesosproto.Call_RECONCILE.Enum(),
		Reconcile: &mesosproto.Call_Reconcile{Tasks: oldTasks},
	})

	if err != nil {
		logrus.WithField("func", "scheduler.reconcile").Debug("Reconcile Error: ", err)
	}
}

// implicitReconcile will ask Mesos which tasks and there state are registert to this framework
func (e *Scheduler) implicitReconcile() {
	var noTasks []*mesosproto.Call_Reconcile_Task
	err := e.Mesos.Call(&mesosproto.Call{
		Type:      mesosproto.Call_RECONCILE.Enum(),
		Reconcile: &mesosproto.Call_Reconcile{Tasks: noTasks},
	})

	if err != nil {
		logrus.WithField("func", "scheduler.implicitReconcile").Debug("Reconcile Error: ", err)
	}
}

func (e *Scheduler) callPluginEvent(event *mesosproto.Event) {
	if e.Config.PluginsEnable {
		for _, p := range e.Config.Plugins {
			symbol, err := p.Lookup("Event")
			if err != nil {
				logrus.WithField("func", "scheduler.callPluginEvent").Error("Error lookup event function in plugin: ", err.Error())
				continue
			}

			eventPluginFunc, ok := symbol.(func(*mesosproto.Event))
			if !ok {
				logrus.WithField("func", "main.callPluginEvent").Error("Error plugin does not have Event function")
				continue
			}

			eventPluginFunc(event)
		}
	}
}
