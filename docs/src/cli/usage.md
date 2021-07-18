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

```

## List all M3s frameworks

```bash
mesos m3s list
ID                                         Active  WebUI                    Name  
2f0fc78c-bf81-4fe0-8720-e27ba217adae-0004  True    http://andreas-pc:10000  m3s   
```

## Get the kubeconfig from the running m3s framework

```bash
mesos m3s kubeconfig 2f0fc78c-bf81-4fe0-8720-e27ba217adae-0003
```

