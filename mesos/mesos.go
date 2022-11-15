package mesos

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net/http"
	"strings"

	api "github.com/AVENTER-UG/mesos-m3s/api"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	"github.com/AVENTER-UG/util/util"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
)

// Scheduler include all the current vars and global config
type Scheduler struct {
	Config    *cfg.Config
	Framework *mesosutil.FrameworkConfig
	Client    *http.Client
	Req       *http.Request
	API       *api.API
}

// Marshaler to serialize Protobuf Message to JSON
var marshaller = jsonpb.Marshaler{
	EnumsAsInts: false,
	Indent:      " ",
	OrigName:    true,
}

// Subscribe to the mesos backend
func Subscribe(cfg *cfg.Config, frm *mesosutil.FrameworkConfig) *Scheduler {
	e := &Scheduler{
		Config:    cfg,
		Framework: frm,
	}

	subscribeCall := &mesosproto.Call{
		FrameworkID: e.Framework.FrameworkInfo.ID,
		Type:        mesosproto.Call_SUBSCRIBE,
		Subscribe: &mesosproto.Call_Subscribe{
			FrameworkInfo: &e.Framework.FrameworkInfo,
		},
	}
	logrus.Debug(subscribeCall)
	body, _ := marshaller.MarshalToString(subscribeCall)
	logrus.Debug(body)
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
		logrus.Error("Mesos Master connection error: ", err.Error())
		return
	}
	defer res.Body.Close()

	reader := bufio.NewReader(res.Body)

	line, _ := reader.ReadString('\n')
	_ = strings.TrimSuffix(line, "\n")

	go e.HeartbeatLoop()

	for {
		// Read line from Mesos
		line, err = reader.ReadString('\n')
		if err != nil {
			logrus.Error("Error to read data from Mesos Master: ", err.Error())
			return
		}
		line = strings.TrimSuffix(line, "\n")
		// Read important data
		var event mesosproto.Event // Event as ProtoBuf
		err := jsonpb.UnmarshalString(line, &event)
		if err != nil {
			logrus.Error("Could not unmarshal Mesos Master data: ", err.Error())
			continue
		}
		logrus.Debug("Subscribe Got: ", event.GetType())

		switch event.Type {
		case mesosproto.Event_SUBSCRIBED:
			logrus.Info("Subscribed")
			logrus.Debug("FrameworkId: ", event.Subscribed.GetFrameworkID())
			e.Framework.FrameworkInfo.ID = event.Subscribed.GetFrameworkID()
			e.Framework.MesosStreamID = res.Header.Get("Mesos-Stream-Id")

			e.Reconcile()
			e.CheckState()
			e.API.SaveFrameworkRedis()
			e.API.SaveConfig()
		case mesosproto.Event_UPDATE:
			e.HandleUpdate(&event)
			// save configuration
			e.API.SaveConfig()
		case mesosproto.Event_OFFERS:
			// Search Failed containers and restart them
			logrus.Debug("Offer Got")
			err = e.HandleOffers(event.Offers)
			if err != nil {
				logrus.Error("Switch Event HandleOffers: ", err)
			}
		}
	}
}

// Generate "num" numbers of random host portnumbers
func (e *Scheduler) getRandomHostPort(num int) uint32 {
	// search two free ports
	for i := e.Framework.PortRangeFrom; i < e.Framework.PortRangeTo; i++ {
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
	keys := e.API.GetAllRedisKeys(e.Framework.FrameworkName + ":*")
	for keys.Next(e.API.Redis.RedisCTX) {
		// get the details of the current running service
		key := e.API.GetRedisKey(keys.Val())

		if e.API.CheckIfNotTask(keys) {
			continue
		}

		task := mesosutil.DecodeTask(key)

		// check if the given port is already in use
		ports := task.Discovery.GetPorts()
		if ports != nil {
			for _, hostport := range ports.GetPorts() {
				if hostport.Number == port {
					logrus.Debug("Port in use: ", port)
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

func (e *Scheduler) addDockerParameter(current []mesosproto.Parameter, newValues mesosproto.Parameter) []mesosproto.Parameter {
	return append(current, newValues)
}

func (e *Scheduler) appendString(current []string, newValues string) []string {
	return append(current, newValues)
}

// Reconcile will reconcile the task states after the framework was restarted
func (e *Scheduler) Reconcile() {
	logrus.Info("Reconcile Tasks")
	var oldTasks []mesosproto.Call_Reconcile_Task
	keys := e.API.GetAllRedisKeys(e.Framework.FrameworkName + ":*")
	for keys.Next(e.API.Redis.RedisCTX) {
		// continue if the key is not a mesos task
		if e.API.CheckIfNotTask(keys) {
			continue
		}

		key := e.API.GetRedisKey(keys.Val())

		task := mesosutil.DecodeTask(key)

		if task.TaskID == "" || task.Agent == "" {
			continue
		}

		oldTasks = append(oldTasks, mesosproto.Call_Reconcile_Task{
			TaskID: mesosproto.TaskID{
				Value: task.TaskID,
			},
			AgentID: &mesosproto.AgentID{
				Value: task.Agent,
			},
		})
		logrus.Debug("Reconcile Task: ", task.TaskID)
	}
	err := mesosutil.Call(&mesosproto.Call{
		Type:      mesosproto.Call_RECONCILE,
		Reconcile: &mesosproto.Call_Reconcile{Tasks: oldTasks},
	})

	if err != nil {
		e.API.ErrorMessage(3, "Reconcile_Error", err.Error())
		logrus.Debug("Reconcile Error: ", err)
	}
}
