# Configuration

| Variable | Default | Description |
| --- | --- | ---- |
| FRAMEWORK_USER        | root | |
| FRAMEWORK_NAME        | | |
| FRAMEWORK_ROLE        | m3s | |
| FRAMEWORK_PORT        | 10000 | |
| FRAMEWORK_HOSTNAME    | ${HOSTNAME} | |
| MESOS_USERNAME        | | |
| MESOS_PASSWORD        | | |
| MESOS_MASTER          | | |
| MESOS_CNI             | | |
| MESOS_SSL             | false | |
| MESOS_PRINCIPAL       | | |
| PORTRANGE_FROM        | 31000 | |
| PORTRANGE_TO          | 32000 | |
| LOGLEVEL              | info | |
| DOCKER_CNI            | bride | |
| DOCKER_SOCK           | | |
| DOCKER_SHM_SIZE       | 30gb | |
| DOMAIN                | .local | |
| AUTH_USERNAME         | | |
| AUTH_PASSWORD         | | |
| K3S_TOKEN             | 123456789 | |
| K3S_CUSTOM_DOMAIN     | cloud.local | |
| K3S_SERVER_STRING     | /usr/local/bin/k3s server --cluster-cidr=10.2.0.0/16 --service-cidr=10.3.0.0/16 --cluster-dns=10.3.0.10  --kube-controller-manager-arg='leader-elect=false' --disable-cloud-controller --kube-scheduler-arg='leader-elect=false' --snapshotter=native --flannel-backend=vxlan | |
|	K3S_SERVER_CPU        | 0.1 | |
|	K3S_SERVER_MEM        | 1200 | |
| K3S_SERVER_CONSTRAINT | | |
| K3S_AGENT_STRING      | /usr/local/bin/k3s agent --snapshotter=native --flannel-backend=vxlan | |
|	K3S_AGENT_CPU         | 0.1 | |
|	K3S_AGENT_MEM         | 1200 | |
| K3S_AGENT_LABELS=[{"key":"traefik.enable","value":"true"},{"key":"traefik.http.routers.m3s.entrypoints","value":"web"},{"key":"traefik.http.routers.m3s.service","value":"m3s-http"},{"key":"traefik.http.routers.m3s.rule","value":"HostRegexp(`example.com`, `{subdomain:[a-z]+}.example.com`)"}] | |

| REDIS_PASSWORD        | | |
| REDIS_SERVER          | 127.0.0.1:6379 | |
| BOOTSTRAP_URL=https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/master/bootstrap/bootstrap.sh
| SKIP_SSL=true
| SSL_CRT_BASE64=<cat server.crt | base64 -w 0>
| SSL_KEY_BASE64=<cat server.key | base64 -w 0>
| ETCD_CONSTRAINT=
| K3S_ETCD_CONSTRAINT=
| K3S_AGENT_COUNT=1
| K3S_AGENT_CONSTRAINT=
| K3S_DOCKER=
|	ETCD_CPU=
|	ETCD_MEM=
|	ETCD_DISK=
| IMAGE_K3S=
| IMAGE_ETCD=
| VOLUME_DRIVER=
| VOLUME_K3S_SERVER=
