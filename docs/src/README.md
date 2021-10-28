# M3s - Apache Mesos Kubernetes Framework

## Introduction

M3s is a Golang based Apache Mesos Framework to run and deploy Kubernetes.

## Requirements


- Apache Mesos min 1.6.0
- Mesos with SSL and Authentication is optional
- Persistent Storage to store Kubernetes data

## Screenshots

### Mesos

![Mesos](vx_images/Mesos.png)


### Kubernetes Dashboard

Get the token and start the proxy.

```bash

kubectl -n kubernetes-dashboard describe secret admin-user-token | grep '^token'
kubectl proxy

```

And then open the browser:

http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/

![Kubernetes](vx_images/Kubernetes.png)


### Traefik

```bash

kubectl port-forward $(kubectl get pods --selector "app.kubernetes.io/name=traefik" --output=name -n kube-system) -n kube-system 9000:9000

```

![Traefik](vx_images/Traefik.png)
