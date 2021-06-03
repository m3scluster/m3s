#!/bin/bash

cat /etc/resolv.conf

apt-get update -y 
apt-get install -y containernetworking-plugins containerd tcpdump curl inetutils-ping iptables fuse-overlayfs procps bash iproute2
mkdir -p /etc/cni/net.d

export INSTALL_K3S_SKIP_ENABLE=true
export INSTALL_K3S_SKIP_START=true
export KUBECONFIG=$MESOS_SANDBOX/kubeconfig.yaml

update-alternatives --set iptables /usr/sbin/iptables-legacy
curl -sfL https://get.k3s.io | sh -
curl https://raw.githubusercontent.com/AVENTER-UG/go-mesos-framework-k3s/master/bootstrap/dashboard_auth.yaml > $MESOS_SANDBOX/dashboard_auth.yaml
curl https://raw.githubusercontent.com/kubernetes/dashboard/v2.2.0/aio/deploy/recommended.yaml > $MESOS_SANDBOX/dashboard.yaml
if [ "$K3SFRAMEWORK_TYPE" = "server" ]
then
  curl -L "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" > /usr/local/bin/kubectl
  curl https://raw.githubusercontent.com/AVENTER-UG/go-mesos-framework-k3s/master/bootstrap/server > $MESOS_SANDBOX/server
  chmod +x /usr/local/bin/kubectl
  chmod +x $MESOS_SANDBOX/server
  exec $MESOS_SANDBOX/server &
fi

echo $1
$1 
