# mesos-m3s
[![Discussion](https://img.shields.io/static/v1?label=&message=Discussion&color=brightgreen)](https://github.com/m3scluster/m3s/discussions)
[![Issues](https://img.shields.io/static/v1?label=&message=Issues&color=brightgreen)](https://github.com/m3scluster/m3s/issues)
[![Chat](https://img.shields.io/static/v1?label=&message=Chat&color=brightgreen)](https://matrix.to/#/#mesos:matrix.aventer.biz?via=matrix.aventer.biz)
[![Docs](https://img.shields.io/static/v1?label=&message=Docs&color=brightgreen)](https://m3scluster.github.io/m3s/)
[![Docker Pulls](https://img.shields.io/docker/pulls/avhost/mesos-m3s)](https://hub.docker.com/repository/docker/avhost/mesos-m3s/)

Mesos Framework to run Kubernetes (K3S)

## Funding

[![](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/donate/?hosted_button_id=H553XE4QJ9GJ8)

## Issues

To open an issue, please use this place: https://github.com/m3scluster/m3s/issues

## Requirements

- Apache Mesos min 1.6.0
- Mesos with SSL and Authentication is optional
- Persistent Storage to keep K3S data (not object storage)
- Redis DB

## Run Framework

The following environment parameters are only a example. All parameters and the default values are documented in
the `init.go` file (real documentation will be coming later). These example assume, that we run mesos-mini.

### Step 1

Run a redis server:

```Bash
docker run --rm --name redis -d -p 6379:6379 redis
```

### Step 2

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

### Step 3 

Before we launch M3s, we create in Docker in dedicated network.

```Bash
docker network create --subnet 10.40.0.0/24 mini
```

### Step 4

Now M3s can be started:

```Bash
./mesos-m3s
```

### Mesos-M3s in real Apache Mesos environments

In real mesos environments, we have to set at least the following environment variables:

```Bash
export MESOS_MASTER="leader.mesos:5050"
export MESOS_USERNAME=""
export MESOS_PASSWORD=""
```

Also the following could be usefull.

```Bash
export REDIS_SERVER="127.0.0.1:6379"
export REDIS_PASSWORD=""
export REDIS_DB="1"
export MESOS_CNI="weave"
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
