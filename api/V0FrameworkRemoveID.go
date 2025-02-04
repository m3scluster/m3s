package api

import (
	cTls "crypto/tls"
	"encoding/json"
	"net/http"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"
	util "github.com/AVENTER-UG/util/util"
	"github.com/sirupsen/logrus"
)

// V0FrameworkRemoveID Cleanup the Framework ID
// example:
// curl -X DELETE http://user:password@127.0.0.1:10000/api/compose/v0/framework
func (e *API) V0FrameworkRemoveID(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("func", "api.V0FrameworkRemoveID").Debug("Remove Framework ID")

	auth := e.CheckAuth(r, w)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	if !auth {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	client := &http.Client{}
	client.Transport = &http.Transport{
		TLSClientConfig: &cTls.Config{InsecureSkipVerify: true},
	}

	protocol := "https"
	if !e.Framework.MesosSSL {
		protocol = "http"
	}

	req, _ := http.NewRequest("GET", protocol+"://"+e.Framework.MesosMasterServer+"/frameworks/?framework_id="+e.Framework.FrameworkInfo.Id.GetValue(), nil)
	req.SetBasicAuth(e.Framework.Username, e.Framework.Password)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		logrus.WithField("func", "api.V0FrameworkRemoveID").Errorf("Mesos Master connection error: %s", err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		logrus.WithField("func", "api.V0FrameworkRemoveID").Errorf("http status code: %s", err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	var framework cfg.MesosFrameworkStatus
	err = json.NewDecoder(res.Body).Decode(&framework)

	if err != nil {
		logrus.WithField("func", "api.V0FrameworkRemoveID").Errorf("Could not decode mesos framework information: %s", err.Error())
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// Search the current framework id. If it's inside, then the framework is
	// still registered and we do not have to remove it
	for _, v := range framework.Frameworks {
		if v.ID == e.Framework.FrameworkInfo.Id.GetValue() {
			return
		}
	}

	logrus.WithField("func", "api.V0FrameworkRemoveID").Infof("Remove unregistered Framework ID: %s", e.Framework.FrameworkInfo.Id.GetValue())

	// Remove framework ID
	e.Framework.FrameworkInfo.Id.Value = util.StringToPointer("")
	e.Redis.SaveFrameworkRedis(e.Framework)
}
