package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// V0ScaleEtcd will scale the k3s agent service
// example:
// curl -X GET http://user:password@127.0.0.1:10000/v0/etcd/scale/{count of instances} -d 'JSON'
func V0ScaleEtcd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := CheckAuth(r, w)

	if vars == nil || !auth {
		return
	}

	d := []byte("nok")

	/*

		if vars["count"] != "" {
			newCount, _ := strconv.Atoi(vars["count"])
			oldCount := config.ETCDMax
			logrus.Debug("V0ScaleEtcd: oldCount: ", oldCount)
			config.ETCDMax = newCount
			i := (newCount - oldCount)
			// change the number to be positiv
			if i < 0 {
				i = i * -1
			}

			// Scale Up
			if newCount > oldCount {
				logrus.Info("Etcd Scale Up ", i)
				revive := &mesosproto.Call{
					Type: mesosproto.Call_REVIVE,
				}
				mesosutil.Call(revive)
			}

			// Scale Down
			if newCount < oldCount {
				logrus.Info("Etcd Scale Down ", i)

				for x := newCount; x < oldCount; x++ {
					task := mesos.StatusEtcd(x)
					if task != nil {
						id := task.Status.TaskID.Value
						ret := mesos.Kill(id)

						logrus.Info("V0TaskKill: ", ret)
						config.ETCDCount--
					}
				}
			}

			d = []byte(strconv.Itoa(newCount - oldCount))
		}

		logrus.Debug("HTTP GET V0ScaleEtcd: ", string(d))
	*/
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Api-Service", "v0")

	w.Write(d)
}
