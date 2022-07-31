package mesos

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	api "github.com/AVENTER-UG/mesos-m3s/api"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	"github.com/AVENTER-UG/util"

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
		logrus.Fatal(err)
	}
	defer res.Body.Close()

	reader := bufio.NewReader(res.Body)

	line, _ := reader.ReadString('\n')
	_ = strings.TrimSuffix(line, "\n")

	go e.HeartbeatLoop()

	for {
		// Read line from Mesos
		line, _ = reader.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")
		// Read important data
		var event mesosproto.Event // Event as ProtoBuf
		err := jsonpb.UnmarshalString(line, &event)
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("Subscribe Got: ", event.GetType())

		switch event.Type {
		case mesosproto.Event_SUBSCRIBED:
			logrus.Info("Subscribed")
			logrus.Debug("FrameworkId: ", event.Subscribed.GetFrameworkID())
			e.Framework.FrameworkInfo.ID = event.Subscribed.GetFrameworkID()
			e.Framework.MesosStreamID = res.Header.Get("Mesos-Stream-Id")
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

// Generate random host portnumber
func (e *Scheduler) getRandomHostPort(num int) uint32 {
	// search two free ports
	for i := e.Framework.PortRangeFrom; i < e.Framework.PortRangeTo; i++ {
		port := uint32(i)
		use := false
		for x := 0; x < num; x++ {
			if e.portInUse(port+uint32(x), "server") || e.portInUse(port+uint32(x), "agent") {
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
func (e *Scheduler) portInUse(port uint32, service string) bool {
	// get all running services
	logrus.Debug("Check if port is in use: ", port, service)
	keys := e.API.GetAllRedisKeys(e.Framework.FrameworkName + ":" + service + ":*")
	for keys.Next(e.API.Redis.RedisCTX) {
		// get the details of the current running service
		key := e.API.GetRedisKey(keys.Val())
		task := mesosutil.DecodeTask(key)

		// check if the given port is already in use
		ports := task.Discovery.GetPorts()
		if ports != nil {
			for _, hostport := range ports.GetPorts() {
				if hostport.Number == port {
					logrus.Debug("Port in use: ", port, service)
					return true
				}
			}
		}
	}
	return false
}

// get network info of task
func (e *Scheduler) getNetworkInfo(taskID string) []mesosproto.NetworkInfo {
	client := &http.Client{}
	// #nosec G402
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: e.Config.SkipSSL},
	}

	protocol := "https"
	if !e.Framework.MesosSSL {
		protocol = "http"
	}
	req, _ := http.NewRequest("POST", protocol+"://"+e.Framework.MesosMasterServer+"/tasks/?task_id="+taskID+"&framework_id="+e.Framework.FrameworkInfo.ID.GetValue(), nil)
	req.Close = true
	req.SetBasicAuth(e.Framework.Username, e.Framework.Password)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		logrus.WithField("func", "getNetworkInfo").Error("Could not connect to agent: ", err.Error())
		return []mesosproto.NetworkInfo{}
	}

	defer res.Body.Close()

	var task cfg.MesosTasks
	err = json.NewDecoder(res.Body).Decode(&task)
	if err != nil {
		logrus.WithField("func", "getAgentInfo").Error("Could not encode json result: ", err.Error())
		return []mesosproto.NetworkInfo{}
	}

	if len(task.Tasks) > 0 {
		for _, status := range task.Tasks[0].Statuses {
			if status.State == "TASK_RUNNING" {
				var netw []mesosproto.NetworkInfo
				netw = append(netw, status.ContainerStatus.NetworkInfos[0])
				// try to resolv the tasks hostname
				if task.Tasks[0].Container.Hostname != nil {
					addr, err := net.LookupIP(*task.Tasks[0].Container.Hostname)
					if err == nil {
						hostNet := []mesosproto.NetworkInfo{{
							IPAddresses: []mesosproto.NetworkInfo_IPAddress{{
								IPAddress: func() *string { x := string(addr[0]); return &x }(),
							}}},
						}
						netw = append(netw, hostNet[0])
					}
				}
				return netw
			}
		}
	}
	return []mesosproto.NetworkInfo{}
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
