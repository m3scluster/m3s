# Containerd sideload configuration


M3s is a high flexible mesos framework. The decision to use a [bootstap script](https://github.com/AVENTER-UG/mesos-m3s/blob/master/bootstrap/bootstrap.sh)
give us the posibility to change the default configuration of K3s. These feature is important if we need access to a private insecure container registry.
The following steps should be a example how to use it.


1) Create a repository with a copy of the official `bootstrap.sh`
2) In the same repository, create a `config.toml.tmpl` file with the following content:

```
[plugins.opt]
  path = "/var/lib/rancher/k3s/agent/containerd"

[plugins.cri]
  stream_server_address = "127.0.0.1"
  stream_server_port = "10010"
  enable_selinux = false
  sandbox_image = "rancher/pause:3.1"

[plugins.cri.containerd]
  disable_snapshot_annotations = true
  snapshotter = "native"

[plugins.cri.cni]
  bin_dir = "/var/lib/rancher/k3s/data/current/bin"
  conf_dir = "/var/lib/rancher/k3s/agent/etc/cni/net.d"

[plugins.cri.containerd.runtimes.runc]
  runtime_type = "io.containerd.runc.v2"

[plugins.cri.registry.mirrors."<MY_INSECURE_REGISTRY>"]
endpoint=https://<MY_INSECURE_REGISTRY>

[plugins.cri.registry.configs."<MY_INSECURE_REGISTRY>".tls]
insecure_skip_verify = true

```

3) In the `K3SFRAMEWORK_TYPE == "agent"` section, we will add the following lines:

```bash
curl http://<MY_GIT_REPO>/m3sbootstrap/config.toml.tmpl > $MESOS_SANDBOX/config.toml.tmpl
mkdir /var/lib/rancher/k3s/containerd
cp $MESOS_SANDBOX/config.toml.tmpl /var/lib/ranger/k3s/containerd/
```

4) Tell the M3s framework to use the custom bootstrap.sh.

```bash
export BOOTSTRAP_URL="http://<MY_GIT_REPO>/m3sbootstrap/bootstrap.sh"
```

5) Done
