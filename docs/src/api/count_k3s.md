# Show count of running Kubernetes Servers/Agents

To get out information about the current running and/or expected Kubernetes Servers/Agents, we can do
the following call.

```bash
curl -X GET -u <username>:<password> http://127.0.0.1:10000/api/m3s/v0/<server|agent>/scale
```
