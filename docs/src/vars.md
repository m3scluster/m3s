# Configuration

The following environment variables are available:


| Variable | Default | Description |
| --- | --- | ---- |
| FRAMEWORK_USER        | root | Framework user used in Apache Mesos |
| FRAMEWORK_NAME        | m3s | Name of the framework in Apache Mesos but also used as Redis prefix |
| FRAMEWORK_ROLE        | m3s | Framework role used in Apache Mesos |
| FRAMEWORK_PORT        | 10000 | Port these framework is listening |
| FRAMEWORK_HOSTNAME    | ${HOSTNAME} | The frameworks hostname |
| MESOS_USERNAME        | | Username to authenticate against Mesos |
| MESOS_PASSWORD        | | Password to authenticate against Mesos |
| MESOS_MASTER          | 127.0.0.1:5050 | Adress of the Mesos Master. If you use mesos-dns, use leader.mesos |
| MESOS_CNI             | | Mesos CNI M3s should use |
| MESOS_SSL             | false | Enable SSL for the communication to the Mesos Master |
| MESOS_PRINCIPAL       | | Mesos Principal |
| PORTRANGE_FROM        | 31000 | Set the portrange, M3s is suposed to use for the container outside of K8 |
| PORTRANGE_TO          | 32000 | |
| LOGLEVEL              | info | Information Level (info, warn, error, debug)|
| DOCKER_CNI            | bride | If we do not use Mesos CNI, we can also use docker network |
| DOCKER_SOCK           | | The docker sock file |
| DOCKER_SHM_SIZE       | 30gb | Size of dockers shared memory |
| DOMAIN                | .local | The domain of the hostnames. As example, if you use weave cni, it would be weave.local |
| AUTH_USERNAME         | | Username to authenticate against these framework |
| AUTH_PASSWORD         | | Password to authenticate against these framework |
| CGROUP_V2             | false | Enable support for CGroupV2 | 
| K3S_TOKEN             | 123456789 | K8 token for the bootstrap |
| K3S_CUSTOM_DOMAIN     | cloud.local | The network Domain we will use for the K3s cni |
| K3S_SERVER_STRING     | /usr/local/bin/k3s server --cluster-cidr=10.2.0.0/16 --service-cidr=10.3.0.0/16 --cluster-dns=10.3.0.10  --kube-controller-manager-arg='leader-elect=false' --disable-cloud-controller --kube-scheduler-arg='leader-elect=false' --snapshotter=native --flannel-backend=vxlan | These is the string we will use to start the K3s server. M3s will add several other parameters. |
|	K3S_SERVER_CPU        | 0.1 | Resources for the K3s Server container |
|	K3S_SERVER_MEM        | 1200 | |
| K3S_SERVER_CONSTRAINT | <hostname> | Tell Mesos to start the K3s server on this hostname |
| K3S_AGENT_STRING      | /usr/local/bin/k3s agent --snapshotter=native --flannel-backend=vxlan | These is the string we will use to start the K3s agent. M3s will add several other parameters. |
|	K3S_AGENT_CPU         | 0.1 | Resources for the K3s Agent container |
|	K3S_AGENT_MEM         | 1200 | |
| K3S_AGENT_LABELS | [{"key":"traefik.enable","value":"true"},{"key":"traefik.http.routers.m3s.entrypoints","value":"web"},{"key":"traefik.http.routers.m3s.service","value":"m3s-http"},{"key":"traefik.http.routers.m3s.rule","value":"HostRegexp(`example.com`, `{subdomain:[a-z]+}.example.com`)"}] | Configure custom labels for the container. In these example, we will use lables for traefik. |
| REDIS_PASSWORD        | | Redis Passwort for authentication |
| REDIS_SERVER          | 127.0.0.1:6379 | Redis Server IP and port |
| BOOTSTRAP_URL | https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/master/bootstrap/bootstrap.sh | Bootstrap Script to pre configure the server/agent container |
| SKIP_SSL | true | Skip SSL Verification |
| SSL_CRT_BASE64 | <cat server.crt | base64 -w 0> | SSL CRT Content as base64 |
| SSL_KEY_BASE64=<cat server.key | base64 -w 0> | SSL Key Content as base64 |
| K3S_DS_CONSTRAINT  | <hostname> | Tell Mesos to start the datastore on this hostname |
| K3S_AGENT_COUNT  | 1 | Amount of running K3s agents |
| K3S_AGENT_CONSTRAINT  | |  Tell Mesos to start the K3s agent on that hostname |
| K3S_DOCKER  | true | Use docker container as K8 runtime |
|	DS_CPU  | Resources for the datastore container |
|	DS_MEM  | | |
|	DS_DISK  | | |
| DS_PORT | 3306 | Datastore Portnumber |
| IMAGE_K3S | avhost/ubuntu-m3s | Ubuntu M3s Docker Image |
| IMAGE_ETCD | bitnami/etcd | Docker Image for Etcd al Datastore |
| IMAGE_MYSQL | mariadb | Docker Image for MaraiDB as Datastore |
| VOLUME_DRIVER | local | Volume driver docker should use to handle the volume |
| VOLUME_K3S_SERVER | /data/k3s/server/ | Volume name to persist the k3s server data |
| VOLUME_DS | /data/k3s/datastore/ | Volume name to persists the datastore data |
| HEARTBEAT_INTERVAL | 15s | Check the state every 'n seconds | 
