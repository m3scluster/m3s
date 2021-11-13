package mesos

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
)

// Service include all the current vars and global config
var config *cfg.Config
var framework *mesosutil.FrameworkConfig

// Marshaler to serialize Protobuf Message to JSON
var marshaller = jsonpb.Marshaler{
	EnumsAsInts: false,
	Indent:      " ",
	OrigName:    true,
}

// SetConfig set the global config
func SetConfig(cfg *cfg.Config, frm *mesosutil.FrameworkConfig) {
	config = cfg
	framework = frm
}

// Subscribe to the mesos backend
func Subscribe() error {

	subscribeCall := &mesosproto.Call{
		FrameworkID: framework.FrameworkInfo.ID,
		Type:        mesosproto.Call_SUBSCRIBE,
		Subscribe: &mesosproto.Call_Subscribe{
			FrameworkInfo: &framework.FrameworkInfo,
		},
	}
	logrus.Debug(subscribeCall)
	body, _ := marshaller.MarshalToString(subscribeCall)
	logrus.Debug(body)
	client := &http.Client{}
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	protocol := "https"
	if !framework.MesosSSL {
		protocol = "http"
	}
	req, _ := http.NewRequest("POST", protocol+"://"+framework.MesosMasterServer+"/api/v1/scheduler", bytes.NewBuffer([]byte(body)))
	req.Close = true
	req.SetBasicAuth(framework.Username, framework.Password)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		logrus.Fatal(err)
	}

	reader := bufio.NewReader(res.Body)

	line, _ := reader.ReadString('\n')
	bytesCount, _ := strconv.Atoi(strings.Trim(line, "\n"))

	// initialstart
	if framework.MesosStreamID == "" {
		StartEtcd("")
	}

	for {
		// Read line from Mesos
		line, _ = reader.ReadString('\n')
		line = strings.Trim(line, "\n")
		// Read important data
		data := line[:bytesCount]
		// Rest data will be bytes of next message
		bytesCount, _ = strconv.Atoi((line[bytesCount:]))
		var event mesosproto.Event // Event as ProtoBuf
		err := jsonpb.UnmarshalString(data, &event)
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("Subscribe Got: ", event.GetType())

		switch event.Type {
		case mesosproto.Event_SUBSCRIBED:
			logrus.Debug(event)
			logrus.Info("Subscribed")
			logrus.Info("FrameworkId: ", event.Subscribed.GetFrameworkID())
			framework.FrameworkInfo.ID = event.Subscribed.GetFrameworkID()
			framework.MesosStreamID = res.Header.Get("Mesos-Stream-Id")
			d, _ := json.Marshal(&framework)
			err := config.RedisClient.Set(config.RedisCTX, "framework", d, 0).Err()
			if err != nil {
				logrus.Error("Framework save config and state into redis Error: ", err)
			}
		case mesosproto.Event_UPDATE:
			logrus.Debug("Update", HandleUpdate(&event))
			// save configuration
			d, _ := json.Marshal(&config)
			err := config.RedisClient.Set(config.RedisCTX, "framework_config", d, 0).Err()
			if err != nil {
				logrus.Error("Framework save config and state into redis Error: ", err)
			}
		case mesosproto.Event_HEARTBEAT:
			Heartbeat()
		case mesosproto.Event_OFFERS:
			// Search Failed containers and restart them
			logrus.Debug("Offer Got")
			err = HandleOffers(event.Offers)
			if err != nil {
				logrus.Error("Switch Event HandleOffers: ", err)
			}
		default:
			logrus.Debug("DEFAULT EVENT: ", event.Offers)
		}
	}
}

// if all Tasks are running, suppress framework offers
func suppressFramework() {
	logrus.Info("Framework Suppress")
	suppress := &mesosproto.Call{
		Type: mesosproto.Call_SUPPRESS,
	}
	err := mesosutil.Call(suppress)
	if err != nil {
		logrus.Error("Supress Framework Call: ")
	}
}

// Generate random host portnumber
func getRandomHostPort() int {
	rand.Seed(time.Now().UnixNano())
	v := rand.Intn(framework.PortRangeTo-framework.PortRangeFrom) + framework.PortRangeFrom
	return v
}
