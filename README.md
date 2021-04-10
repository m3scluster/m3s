# go-mesos-framework-k3s

[![Donate](https://img.shields.io/liberapay/receives/AVENTER.svg?logo=liberapay)](https://liberapay.com/mesos)
[![Support Chat](https://img.shields.io/static/v1?label=Chat&message=Support&color=brightgreen)](https://riot.im/app/#/room/#support:matrix.aventer.biz)
[![Community](https://img.shields.io/static/v1?label=Community&message=Talk&color=brightgreen)](https://community.aventer.biz/post/46)

Dies ist ein K3S Framework für Apache Mesos

## Voraussetzung

- Apache Mesos ab 1.6.0
- Mesos mit SSL und Authentication ist Optional
- Persistent Storage

## Framework starten

Mit den folgenden Umgebungsvariablen kann das Framework konfiguriert werden. Nach dem Starten wird es sich an den Mesos Master anmelden und erscheint als "k3sframework" in der Mesos UI. Sobald sich das Framework erfolgreich initialisiert hat, startet es Zookeeper und anschließend K3S.



```Bash
export FRAMEWORK_USER="root"
export FRAMEWORK_NAME="k3sframework"
export FRAMEWORK_PORT="10000"
export FRAMEWORK_ROLE="k3s"
export FRAMEWORK_STATEFILE_PATH="/tmp"
export MESOS_PRINCIPAL="<mesos_principal>"
export MESOS_USERNAME="<mesos_user>"
export MESOS_PASSWORD="<mesos_password>"
export MESOS_MASTER="<mesos_master_server>:5050"
export MESOS_CNI="weave"
export LOGLEVEL="DEBUG"
export DOMAIN="weave.local"
export K3S_SERVER_COUNT=1
export K3S_AGENT_COUNT=1
export ETCD_COUNT=1
export RES_CPU=0.1
export RES_MEM=1200
export AUTH_PASSWORD="password"
export AUTH_USERNAME="user"
export MESOS_SSL="true"
export K3S_CUSTOM_DOMAIN=""
export IMAGE_K3S="rancher/k3s:v1.20.0-k3s2"
export IMAGE_ETCD=bitnami/etcd:latest"
export VOLUME_DRIVER="local"
export VOLUME_K3S_SERVER="/tmp/k3s1"

go run init.go app.go
```

![K3S Framework in Mesos](k3s_mesos.gif)

Faellt ein Container aus, wird dieser neugestartet.

![K3S Framework in Mesos](k3s_mesos1.gif)

## Task Status Abfragen

Um den Status eines Tasks über das Framework abzufragen, folgendes Kommando verwenden:

```Bash
curl -X GET 127.0.0.1:10000/v0/container/<taskId> -d 'JSON'  | jq
```

## Fehlende K3S oder Agentr Starten

Sollte aus bestimmten Gründen der Healthcheck Status im Framework nicht mit der Realität übereinstimmen, kann über den nachfolgenden Aufruf erzwungen werden die fehlenden Container zu starten.

```Bash
curl -X GET 127.0.0.1:10000/v0/<server|agent>/reflate -d 'JSON'
```

## K3S Server oder Agent skalieren

Um K3S oder Zookeeper im Betrieb zu skalieren, wird dem Framework die zu laufende Anzahl an Containern angegeben. Soll der Zookeeper also drei mal laufen, muss als <count> eine "3" angegeben werden. Beim Scaledown wird der zuletzt hinzugefügte Container entfernt.

```Bash
curl -X GET 127.0.0.1:10000/v0/<server|agent>/scale/<count> -d 'JSON'
```

## Task killen

Sollte es notwendig sein einen Task zu benden, erfolgt dies mit dem nachfolgenden aufruf. Der beendete Container wird nicht automatisch neu gestartet.

```Bash
curl -X GET 127.0.0.1:10000/v0/task/kill/<taskId> -d 'JSON'
```
