#!/bin/bash

cat /etc/resolv.conf

apt-get update -y
apt-get install -y containerd dnsmasq containernetworking-plugins tcpdump curl inetutils-ping iptables fuse-overlayfs procps bash iproute2 dnsutils net-tools
mkdir -p /etc/cni/net.d

export KUBERNETES_VERSION=v1.21.1
export INSTALL_K3S_VERSION=$KUBERNETES_VERSION+k3s1
export INSTALL_K3S_SKIP_ENABLE=true
export INSTALL_K3S_SKIP_START=true
export KUBECONFIG=$MESOS_SANDBOX/kubeconfig.yaml

update-alternatives --set iptables /usr/sbin/iptables-legacy
curl -sfL https://get.k3s.io | sh -
curl https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/master/bootstrap/dashboard_auth.yaml > $MESOS_SANDBOX/dashboard_auth.yaml
curl https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/master/bootstrap/dashboard_traefik.yaml > $MESOS_SANDBOX/dashboard_traefik.yaml
curl https://raw.githubusercontent.com/kubernetes/dashboard/v2.2.0/aio/deploy/recommended.yaml > $MESOS_SANDBOX/dashboard.yaml
if [[ "$K3SFRAMEWORK_TYPE" == "server" ]]
then
  curl -L http://dl.k8s.io/release/$KUBERNETES_VERSION/bin/linux/amd64/kubectl > $MESOS_SANDBOX/kubectl
  curl https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/master/bootstrap/server > $MESOS_SANDBOX/server
  chmod +x $MESOS_SANDBOX/kubectl
  chmod +x $MESOS_SANDBOX/server
  exec $MESOS_SANDBOX/server &
fi
if [[ "$K3SFRAMEWORK_TYPE" == "agent" ]]
then
  echo "These place you can use to manipulate the configuration of containerd (as example)."
fi


echo $1
$1
