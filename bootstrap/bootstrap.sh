#!/bin/bash

## Example of how to set a custom upstream DNS
echo "nameserver 8.8.8.8" > /etc/k3sresolv.conf

apt-get update -y
mkdir -p /etc/cni/net.d

export INSTALL_K3S_VERSION=$KUBERNETES_VERSION+k3s1
export INSTALL_K3S_SKIP_ENABLE=true
export INSTALL_K3S_SKIP_START=true
export K3S_RESOLV_CONF=/etc/k3sresolv.conf
export BRANCH=master
export ARCH=`dpkg --print-architecture`

## Export json as environment variables
## example: MESOS_SANDBOX_VAR='{ "CUSTOMER":"test-ltd" }'
## echo $CUSTOMER >> test-ltd
for s in $(echo $MESOS_SANDBOX_VAR | jq -r "to_entries|map(\"\(.key)=\(.value|tostring)\")|.[]" ); do
  export $s
done

## dockerd is a part of the uses avhost/ubuntu-m3s:focal docker image
exec /usr/local/bin/dockerd &
sleep 20
exec /usr/bin/cri-dockerd --network-plugin=cni --cni-bin-dir=/usr/lib/cni --cni-conf-dir=/var/lib/rancher/k3s/agent/etc/cni/net.d/ &

curl https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/${BRANCH}/bootstrap/entrypoint-cgroupv2.sh > $MESOS_SANDBOX/entrypoint-cgroupv2.sh

chmod +x $MESOS_SANDBOX/entrypoint-cgroupv2.sh
$MESOS_SANDBOX/entrypoint-cgroupv2.sh

curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=${INSTALL_K3S_VERSION} INSTALL_K3S_SKIP_ENABLE=${INSTALL_K3S_SKIP_ENABLE=$} INSTALL_K3S_SKIP_START=${INSTALL_K3S_SKIP_START} sh -s - --docker
curl https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/${BRANCH}/bootstrap/default.yaml > $MESOS_SANDBOX/default.yaml
if [[ "$K3SFRAMEWORK_TYPE" == "server" ]]
then
  curl -L http://dl.k8s.io/release/$KUBERNETES_VERSION/bin/linux/${ARCH}/kubectl > $MESOS_SANDBOX/kubectl
  curl https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/${BRANCH}/bootstrap/controller.${ARCH} > $MESOS_SANDBOX/controller
  chmod +x $MESOS_SANDBOX/kubectl
  chmod +x $MESOS_SANDBOX/controller
  exec $MESOS_SANDBOX/controller &
fi
if [[ "$K3SFRAMEWORK_TYPE" == "agent" ]]
then
  echo "These place you can use to manipulate the configuration of containerd (as example)."
fi


echo $1
$1
