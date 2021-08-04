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
