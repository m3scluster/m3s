# M3s CLI Installation for Mesos-CLI

The installation of the M3s plugin for mesos-cli is done in few steps.

First, edit the mesos-cli config file.

```bash

vim .mesos/config.toml

```

Add the absolute path of the plugin into the plugin array:

```bash

# The `plugins` array lists the absolute paths of the
# plugins you want to add to the CLI.
plugins = [
  "/example/go-mesos-framework-k3s/mesos_cli/m3s"
]

[m3s]
  principal = "<framework username>"
  secret = "<framework password>"

```

Now we will see the M3s plugin in mesos cli:

```bash

mesos help
Mesos CLI

Usage:
  mesos (-h | --help)
  mesos --version
  mesos <command> [<args>...]

Options:
  -h --help  Show this screen.
  --version  Show version info.

Commands:
  agent   Interacts with the Mesos agents
  config  Interacts with the Mesos CLI configuration file
  m3s     Interacts with the Kubernetes Framework M3s
  task    Interacts with the tasks running in a Mesos cluster

```
