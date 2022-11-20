# mesos-m3s

[![Chat](https://img.shields.io/static/v1?label=Chat&message=Support&color=brightgreen)](https://matrix.to/#/#mesosk3s:matrix.aventer.biz?via=matrix.aventer.biz)
[![Docs](https://img.shields.io/static/v1?label=Docs&message=Support&color=brightgreen)](https://aventer-ug.github.io/mesos-m3s/index.html)

Mesos Framework to run Kubernetes (K3S)

## Requirements

- Apache Mesos min 1.6.0
- Mesos with SSL and Authentication is optional
- Persistent Storage to keep K3S data (not object storage)
- Redis DB

## Run Framework

The following environment parameters are only a example. All parameters and der default values are documented in the `init.go` file (real documentation will be coming later)

### Step 1

M3s needs some parameters to connect to Mesos. The following serve only as an example.

```Bash
export MESOS_SSL=false
export DOCKER_CNI=mini
export LOGLEVEL=DEBUG
export AUTH_USERNAME=user
export AUTH_PASSWORD=password
export VOLUME_K3S_SERVER=local_k3sserver
export K3S_TOKEN=df54383b5659b9280aa1e73e60ef78fc
export DOMAIN=.mini
export BOOTSTRAP_URL=https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/master/bootstrap/bootstrap.sh
export K3S_AGENT_LABELS=[{"key":"traefik.enable","value":"true"},{"key":"traefik.http.routers.m3s.entrypoints","value":"web"},{"key":"traefik.http.routers.m3s.service","value":"m3s-http"},{"key":"traefik.http.routers.m3s.rule","value":"HostRegexp(`example.com`, `{subdomain:[a-z]+}.example.com`)"}]
```

The variable K3S_AGENT_LABELS gives the possibility to create labels for Traefik or other load balancers connected to mesos. In the example given here are labels for our Traefik Provider.

### Step 2

Before we launch M3s, we create in Docker in dedicated network.

```Bash
docker network create --subnet 10.40.0.0/24 mini
```

### Step 3

Now M3s can be started:

```Bash
./mesos-m3s
```

# Screenshots

## Access Kubernetes Dashboard

```bash
kubectl -n kubernetes-dashboard describe secret admin-user-token | grep '^token'
kubectl proxy
```

http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services


![image_2021-05-01-15-09-30](vx_images/image_2021-05-01-15-09-30.png)

## Apache Mesos running K3S Framework

![image_2021-05-01-15-10-54](vx_images/image_2021-05-01-15-10-54.png)

## Access Traefik Dashboard

```bash
kubectl port-forward $(kubectl get pods --selector "app.kubernetes.io/name=traefik" --output=name -n kube-system) -n kube-system 9000:9000
```

http://127.0.0.1:9000/dashboard/


![image_2021-06-13-17-15-45](vx_images/image_2021-06-13-17-15-45.png)
