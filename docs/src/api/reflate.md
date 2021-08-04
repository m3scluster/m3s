# Reflate Kubernetes Server or Agent

To reflate Kubernetes services, we have to call the following API.

```bash
curl -X GET -u <username>:<password> http://127.0.0.1:10000/v0/<server|agent>/reflate -d 'JSON'
```
