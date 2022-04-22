# Scale up/down Kubernetes Service

To scale Kubernetes services, we have to call the following API. `<count>` is the number of all running processes.

```bash
curl -X GET -u <username>:<password> http://127.0.0.1:10000/api/m3s/v0/<server|agent|etcd>/scale/<count>
```
