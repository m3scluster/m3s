# Changelog

## dev
- Update Mesos Protobuf to the current Mesos Version
- Add custom MesosCNI via env

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
