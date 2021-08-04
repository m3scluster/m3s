# Kubectl Usage with M3s

There are a lot of website abotu "how to use kubectl". Therefore we will only describe how to use it together with M3s.

First of all, we need the kubernetes config. To get it, we have to use the mesos-cli. M3s can run multiple times, thats
why we have to choose the right M3s framework.

```bash

mesos m3s list

ID                                         Active  WebUI                    Name
2f0fc78c-bf81-4fe0-8720-e27ba217adae-0004  True    http://m3sframeworkserver:10000  m3s

```

With the framework ID, we can get out the kubernetes config.

```bash

mesos m3s kubeconfig 2f0fc78c-bf81-4fe0-8720-e27ba217adae-0004 > .kube/config

```

Let us have a look into the config file.

```bash

cat .kube/config

apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: <DATA>
    server: https://<MESOS-WORKER>:31862
  name: default
contexts:
- context:
    cluster: default
    user: default
  name: default
current-context: default
kind: Config
preferences: {}

```

The config file include the certificate to authenticate against the kubernetes cluster, and also the server URL.
Like you see, the server URL is one of the Mesos Agents with a dynamic port.

With these config file, we can use kubernetes as we know it.

As example:


```bash

kubectl get nodes

NAME                             STATUS   ROLES                  AGE   VERSION
k3sagent0.weave.local-b5d0de31   Ready    <none>                 40m   v1.21.1+k3s1
k3sserver.weave.local            Ready    control-plane,master   42m   v1.21.1+k3s1

```

In our case, we use weaveworks as container network under Mesos, thats why the names of the kubernetes nodes
contain a "weave.local" domain. Basically, M3s can run under every container network. But it's important that
the names of the containers are resolvable.

Now let us have a look which Kubernetes services are running by default.


```bash

kubectl get svc --all-namespaces

NAMESPACE              NAME                        TYPE           CLUSTER-IP     EXTERNAL-IP               PORT(S)                      AGE
default                kubernetes                  ClusterIP      10.3.0.1       <none>                    443/TCP                      47m
kube-system            kube-dns                    ClusterIP      10.3.0.10      <none>                    53/UDP,53/TCP,9153/TCP       46m
kube-system            metrics-server              ClusterIP      10.3.198.101   <none>                    443/TCP                      46m
kube-system            traefik                     LoadBalancer   10.3.48.210    10.1.1.10,10.1.1.11       80:31455/TCP,443:31347/TCP   44m
kubernetes-dashboard   dashboard-metrics-scraper   ClusterIP      10.3.201.186   <none>                    8000/TCP                     47m
kubernetes-dashboard   kubernetes-dashboard        ClusterIP      10.3.43.87     <none>                    443/TCP                      47m

```
