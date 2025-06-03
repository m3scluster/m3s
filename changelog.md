# Changelog

## master

- ADD: API Endpoint to cleanup the Framework ID. That will force a resubscription under
       a new Framework ID.
- FIX: CGroupsV2 Support
- ADD: Enable Disk allocation through the env variable `RESTRICT_DISK_ALLOCATION`
			 The limits has to be configured through the env variables `K3S_SERVER_DISK_LIMIT` and `K3S_AGENT_DISK_LIMIT`
- ADD: Custom runtime for docker engine through the env variable `CUSTOM_DOCKER_RUNTIME`
- ADD: Support for K3s "Distributed Registry mirror". It can be enabled through
			 the boolean env variable `ENABLE_REGISTRY_MIRROR`.
- CHANGE: Update k3s to version 1.31.5

## v0.5.2

- SECURITY Updates

## v0.5.1

- SECURITY Updates

## v0.5.0

- Add implicit reconcile to remove unknown Mesos Tasks
- Add limitation of memory swap (DOCKER_MEMORY_SWAP, default value "1000" MB) the container can use
- Add API and cli to shutdown the K8 cluster (does not delete persistence volumes)
- Add API and cli to start the K8 cluster after a shutdown
- Add API and cli to restart the K8 cluster 
- Change kubeapi to a fix hostport. It would ever be the first port of the framework portrange
- Fix port detection if the framework is running as container at the same servers like the datastore and K8 manager
- Add support for Server labels
- Add REFUSE_OFFERS to tell mesos it does not have to send new offers during the next seconds (default: 120.0). That will give other 
  frameworks the chance to get offers more quickly.
- Bootstrap: Combine dashboard.yaml, dashboard_auth.yaml and dashboard_traefik.yaml into default.yaml. These file is the place to customize 
  k3s during the bootstrap process.
- Rewrite bootstrap server as Kubernetes Controller to simplify the bootstrap process and optimize cluster health checks.
- Add Kubernetes taint to prevent pods to run on the Kubernetes management node. With K3S_ENABLE_TAINT you can enable(true and default)/disable(disable) these feature.
- Change the mesos cli plugin to avmesos-cli.
- Fix ClusterRestart API [#14](https://github.com/AVENTER-UG/mesos-m3s/pull/14) (thanks to [@itsoksarvesh](https://github.com/itsoksarvesh)).
- Add API to disclosure API capabilities [#16](https://github.com/AVENTER-UG/mesos-m3s/pull/16) (thanks to [@itsoksarvesh](https://github.com/itsoksarvesh)).
- Update K3s to Version 1.25.2
- Move environment variable KUBERNETES_VERSION from bootstrap file into m3s framework.
- Add Timezone support via env variable `TZ` [#17](https://github.com/AVENTER-UG/mesos-m3s/pull/17) (thanks to [@itsoksarvesh](https://github.com/itsoksarvesh)).
- Set Kubernetes agent nodes to unscheduled until all agent nodes are not ready.
- Update golang libraries.
- Add API to restart agent, datastore and server and update CPU and MEM Ressources [#18](https://github.com/AVENTER-UG/mesos-m3s/pull/18) (thanks to [@itsoksarvesh](https://github.com/itsoksarvesh)).
- Change datastore healtcheck to Mesos internal check mechanism.
- Add support to develope plugins for m3s. As example we have add two plugins.
- Add Kafka plugin to stream Mesos event messages to kafka. To enable plugin support, set `M3S_PLUGIN=true`
- Move kubernetes controller into the framework to optimise Kubernetes node handling and healthchecks
- Update taskid label of Kubernetes controll-plane node after restart
- Fix and optimise mesos offerhandling 
- Add support for a TCP port beside HTTP and HTTPS at the Kubernetes agent. It will be configures and
  enabled with the env variable `K3S_AGENT_TCP_PORT`.
- Migrate from gogo protobuf to golang protobuf. Update mesos proto files.
- Add Conditional Offer Constraints, Custom Env Variables, Node Taints set from K3s Arguments
- Fix unneeded Code Removed
- Disable metrics server for m3s controller to fix port conflict.
- Add Set Task Limits
- Fix Status M3s API Panics
- FIX: Update go modudles
- FIX: Update K8S modules + bootstrap for cgroupsv2

## v0.4.2

- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/15
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/16
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/11

## v0.4.1

- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/9
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/8
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/7
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/6
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/5
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/4
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/3
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/2
- FIX-SECURITY: https://github.com/AVENTER-UG/mesos-m3s/security/dependabot/1

## v0.4.0

- Fix docker user defined network.
- Add redis reconnect after unhealthy ping. 
- Add mysql support as datastore endpoint
- Add healthcheck for datastore
- Add resubscription after the connection to mesos master is lost.
- Add container volume for the datastore to persist data
- Add hashicorp vault support for environment variables of the framework
- Add support for ssl and authentication of the bootstrap server
- Add example of how to add custome upstream dns for coredns
- Update to k3s 1.24.x
- Fix scale up performance
- Add support for cri-docker (enable with K3S_DOCKER=true)
- Add support for MySQL TLS datastore communication
- Add support for CGroupV2 (bool env CGROUP_V2, default false)
- Change cli to support multicluster    
- Add ulimit docker parameter (DOCKER_ULIMIT)
- Change DOCKER_SHM_SIZE to K3S_CONTAINER_DISK
- Add mesos reconcile loop to periodically sync state with mesos
- Fix GetNetworkInfo resolv hostname


## v0.3.0

- Add Redis Authentication Support
- Bootstrap: Add update call to update the bootstrap server.
- Add update api call to update the bootstrap server.
- Fix hostname constraint method. The server/agent should never run
  on a other worker as defined in the contrains variable.
- Fix mesos-cli to determine framework uuid by name.
- Move statefile to Redis DB
- Add TLS Server Support (env variable SSL_CRT_BASE64, SSL_KEY_BASE64)
- Change DB items framework and framework_config to be saved with the
  frameworkName as prefix.
- Fix M3SStatus and scalet
- Optimize framework suppress
- Check if port is already in use
- Add version as flag to bootstrap server and m3s
- IMPORTANT!!! Change all API Calls URL's.
- Add Docker Network (`docker network create`) support. The configuration variable ist "DOCKER_CNI".
- Seperate K3S Server and Agent Memory and CPU resource definition.
- Change K3S to use docker engine
- Add DOCKER_SHM_SIZE variable to configure shm-size.
- Increase etcd election timeout

## v0.2.0

- Add status information about the M3s services.
- Mesos-CLI: Get out the M3s status information.
- Add Mesos Label support to (as example) control traefik
- Add possibility to run the K8 Server and Agent on a specified hostname
- Fix status information of all m3s services
- Mesos-CLI: Resolve framework ID also by name
- Add configurable FrameworkHostname
- Add FrameworkName to Port Label
- Add Kubernetes Status to Framework, bootstrap Server and Mesos-CLI
- Add the possibility to set custom environment variables for the bootstrap.sh
- Remove unneeded API calls
- Add API call to get the current count of running and expected Kubernetes Server/Agents.

## v0.1.0

- Update Mesos Protobuf to the current Mesos Version
- Add custom MesosCNI via env
- Fix snapshooter, flannel and cgroups error
- Fix server/agent mismatch during shutdown
- Removes via api killed tasks from config. Same with scale down agents/servers.
- Add support of external etcd service
- Add configurable docker socket
- Add scale up/down from etcd
- Change container engine to to containerd
- Enable persistent storage for K3S server
- Add wildcard certificate for K3S to prevent certificate error during K3S server restart
- Enable Traefik and ServiceLB
- Add startup order
- Fix communication between pods via different mesos agents
- Add fetch custom bootstrap script to init kubernetes
- Change Mesos TASK ID's
- Change the K3S Image to plain ubuntu. The bootstrap will install K3S.
- Add Dashbaord configuration
- Add K3S Framework API Server into the Container
- Add API to get Kubeconfig
- Add Kubernetes server check to scheduler the agent only if the server is running
- Add automatic Kubernetes dashboard deploy to the API server
- Add seperated memory and CPU resources for ETCD and K3S
- Add Heatbeat for ETCD and K3S services
- Add automatic Traefik dashboard deploy
- Add mesos-cli plugin
- Mesos-CLI: Show all running M3s frameworks
- Mesos-CLI: Get kubeconfig from a selected running m3s
- Add supress offers if every services are running

## v0.0.4

- Add volume driver support

## v0.0.3

- Add persistent storage Support
- Change to go mod

## v0.0.2

- Add start mesos container
- Add start command
- Add persist framework info therefore the framework know what to do after a crash
- Add save task state
- Add start k3s
- Add persist task state
- Add multinode support
- Add container monitor to restart if a container failed
- Add RestCall to reflate missing k3s processes, for the case the monitoring can not find the problem
- Add RestCall scale up and down
- Add RestCall to kill a task
- Add Authentication
- Add Support to configure (non)SSL Support to Mesos
- Add Custom Domain to match external DNS (like Consul)
- Add Service Name ENV Variable to match external DNS (like Consule)
- Add Call_Suppress to tell mesos it does not send us offers until we ask
- Add default values for some env
- Add custom image name via env
- Add schedule start order
