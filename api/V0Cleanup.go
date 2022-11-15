package api

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

var cleanupLock bool

// V0Cleanup will cleanup notready nodes
// example:
// curl -X GET 127.0.0.1:10000/v0/agent/cleanup'
func (e *API) V0Cleanup(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0Cleanup").Debug("Cleanup notready nodes")

	if !e.CheckAuth(r, w) {
		return
	}
	e.CleanupNodes()
}

// CleanupNodes - To cleanup notready nodes
func (e *API) CleanupNodes() {
	if e.Config.K3SServerHostname == "" {
		return
	}

	client := &http.Client{}
	// #nosec G402
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: e.Config.SkipSSL},
	}
	req, _ := http.NewRequest("GET", e.BootstrapProtocol+"://"+e.Config.K3SServerHostname+":"+strconv.Itoa(e.Config.K3SServerContainerPort)+"/api/m3s/bootstrap/v0/clean", nil)
	req.SetBasicAuth(e.Config.BootstrapCredentials.Username, e.Config.BootstrapCredentials.Password)
	req.Close = true
	res, err := client.Do(req)

	if err != nil {
		logrus.WithField("func", "api.V0Cleanup").Error("Could not call bootstrap api: ", err, res)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.WithField("func", "api.V0Cleanup").Error("Call bootstrap api retrun not 200: ", err, res)
		return
	}
}

// ScheduleCleanup - Schedule cleanup notready agents
func (e *API) ScheduleCleanup() {
	if cleanupLock {
		return
	}
	logrus.WithField("func", "api.ScheduleCleanup").Debug("Schedule Cleanup")
	cleanupLock = true

	select {
	case <-time.After(e.Config.CleanupLoopTime):
		e.CleanupNodes()
	}
	cleanupLock = false
}
