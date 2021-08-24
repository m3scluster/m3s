# M3s CLI Usage

The M3s framework does support the new version of mesos-cli.


The following parameters are currently supported:

```bash

mesos m3s

Interacts with the Kubernetes Framework M3s

Usage:
  mesos m3s (-h | --help)
  mesos m3s --version
  mesos m3s <command> (-h | --help)
  mesos m3s [options] <command> [<args>...]

Options:
  -h --help  Show this screen.
  --version  Show version info.

Commands:
  kubeconfig  Get kubernetes configuration file
  list        Show list of running M3s frameworks
  scale       Scale up/down the Manager or Agent of Kubernetes
  status      Get out live status information
  version     Get the version number of Kubernetes
```

## List all M3s frameworks

```bash

mesos m3s list

ID                                         Active  WebUI                    Name
2f0fc78c-bf81-4fe0-8720-e27ba217adae-0004  True    http://andreas-pc:10000  m3s

```

## Get the kubeconfig from the running m3s framework

```bash

mesos m3s kubeconfig 2f0fc78c-bf81-4fe0-8720-e27ba217adae-0004

```

## Scale up/down Kubernetes services


We can scale up/down several Kubernetes services.

```bash

mesos m3s scale
Scale up/down the Manager or Agent of Kubernetes

Usage:
  mesos m3s scale (-h | --help)
  mesos m3s scale --version
  mesos m3s scale [options] <framework-id> <count>

Options:
  -a --agent    Scale up/down Kubernetes agents
  -e --etcd     Scale up/down etcd
  -h --help     Show this screen.
  -m --manager  Scale up/down Kubernetes manager


```

The "count" is the number of how many instances of the selected service should run.

As example:

```bash

 mesos m3s scale --agent 2f0fc78c-bf81-4fe0-8720-e27ba217adae-0004 2

```

## M3s Status overview


The status command support two different flags.

```bash

mesos m3s status
Get out live status information

Usage:
  mesos m3s status (-h | --help)
  mesos m3s status --version
  mesos m3s status [options] <framework-id>

Options:
  -h --help        Show this screen.
  -k --kubernetes  Give out the Kubernetes status.
  -m --m3s         Give out the M3s status.

Description:
  Get out live status information

```

`--kubernetes` (in developing) will give out the stats of the kubernetes environment.

`--m3s` Show the current status of the M3s services.

```bash

mesos m3s status -m 2f0fc78c-bf81-4fe0-8720-e27ba217adae-0004
{"Server":"TASK_RUNNING","Agent":"TASK_RUNNING","API":"ok","Etcd":"TASK_RUNNING"}


```
