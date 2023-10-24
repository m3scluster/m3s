# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
The m3s plugin.
"""

import toml
import pprint

import urllib3

from avmesos import cli
from avmesos.cli.exceptions import CLIException
from avmesos.cli.plugins import PluginBase
from avmesos.cli.util import Table
from avmesos.cli.mesos import get_frameworks, get_framework_address
from avmesos.cli import http


PLUGIN_NAME = "m3s"
PLUGIN_CLASS = "M3s"

VERSION = "0.1.0"

SHORT_HELP = "Interacts with the Kubernetes Framework M3s"


class Config():

    def __init__(self, main):
        """
        Get authentication header for the framework
        """

        self.main = main

        try:
            data = toml.load(self.main.config.path)
        except Exception as exception:
            raise CLIException(
                "Error loading config file as TOML: {error}".format(
                    error=exception)
            ) from exception

        self.data = data["m3s"].get(self.main.framework_name)

    def principal(self):
        """
        Return the principal in the configuration file
        """
        return self.data.get("principal")

    def secret(self):
        """
        Return the secret in the configuration file
        """

        return self.data.get("secret")
    
    def ssl_verify(self, default=False):
        """
        Return if the ssl certificate should be verified
        """
        ssl_verify = self.data.get("ssl_verify", default)
        if not isinstance(ssl_verify, bool):
            raise CLIException("The 'ssl_verify' field must be True/False")

        return ssl_verify

    # pylint: disable=no-self-use
    def agent_timeout(self, default=5):
        """
        Return the connection timeout of the agent
        """

        return default


class M3s(PluginBase):
    """
    The m3s plugin.
    """

    COMMANDS = {
        "kubeconfig": {
            "arguments": ["<framework-name>"],
            "flags": {},
            "short_help": "Get kubernetes configuration file",
            "long_help": "Get kubernetes configuration file",
        },
        "list": {
            "arguments": [],
            "flags": {
                "-a --all": "list all M3s frameworks, not only running [default: False]"
            },
            "short_help": "Show list of running M3s frameworks",
            "long_help": "Show list of running M3s frameworks",
        },
        "version": {
            "arguments": ["<framework-name>"],
            "flags": {},
            "short_help": "Get the version number of Kubernetes",
            "long_help": "Get the version number of Kubernetes",
        },
        "status": {
            "arguments": ["<framework-name>"],
            "flags": {
                "-m --m3s": "Give out the M3s status.",
                "-k --kubernetes": "Give out the Kubernetes status.",
            },
            "short_help": "Get out live status information",
            "long_help": "Get out live status information",
        },
        "scale": {
            "arguments": ["<framework-name>", "<count>"],
            "flags": {
                "-a --agent": "Scale up/down Kubernetes agents",
                "-e --etcd": "Scale up/down etcd",
            },
            "short_help": "Scale up/down the Manager or Agent of Kubernetes",
            "long_help": "Scale up/down the Manager or Agent of Kubernetes",
        },
        "cluster": {
            "arguments": ["<framework-name>", "<operations>"],
            "flags": {},
            "short_help": "Control the Kubernetes cluster",
            "long_help": "Control the Kubernetes cluster\noperations:\n\tstop - stop the Kubernetes cluster\n\tstart - start the Kubernetes cluster\n\trestart - restart the Kubernetes cluster\n",
        },
    }

    def cluster(self, argv):
        """
        Control the Kubernetes cluster
        """

        try:
            master = self.config.master()
            config = self.config
            # pylint: disable=attribute-defined-outside-init
            self.framework_name = argv["<framework-name>"]
            self.m3sconfig = Config(self)
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(
                    error=exception)
            ) from exception

        if argv["<operations>"] == "stop":
            print("Shutdown Kubernetes Cluster")

            framework_address = get_framework_address(
                self.get_framework_id(argv), master, config
            )
            data = self.write_endpoint(
                framework_address,
                "/api/m3s/v0/cluster/shutdown",
                self,
                "PUT"
            )
            print(data)
        if argv["<operations>"] == "start":
            print("Start Kubernetes Cluster")

            framework_address = get_framework_address(
                self.get_framework_id(argv), master, config
            )
            data = self.write_endpoint(
                framework_address,
                "/api/m3s/v0/cluster/start",
                self,
                "PUT"
            )
            print(data)
        if argv["<operations>"] == "restart":
            print("Restart Kubernetes Cluster")

            framework_address = get_framework_address(
                self.get_framework_id(argv), master, config
            )
            data = self.write_endpoint(
                framework_address,
                "/api/m3s/v0/cluster/restart",
                self,
                "PUT"
            )
            print(data)

    def scale(self, argv):
        """
        Scale up/down Kubernetes Manager or Agents
        """

        try:
            master = self.config.master()
            config = self.config
            # pylint: disable=attribute-defined-outside-init
            self.framework_name = argv["<framework-name>"]
            self.m3sconfig = Config(self)
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(
                    error=exception)
            ) from exception

        service = None
        if argv["--agent"]:
            service = "agent"
        elif argv["--etcd"]:
            service = "etcd"

        if service is not None:
            print("Scale up/down " + service)

            framework_address = get_framework_address(
                self.get_framework_id(argv), master, config
            )
            data = http.read_endpoint(
                framework_address,
                "/api/m3s/v0/" + service + "/scale/" + argv["<count>"],
                self,
            )
            print(data)
        else:
            print("Nothing to scale")

    def kubeconfig(self, argv):
        """
        Load the kubeconfig file from the kubernetes api server
        """

        try:
            master = self.config.master()
            config = self.config
            # pylint: disable=attribute-defined-outside-init
            self.framework_name = argv["<framework-name>"]
            self.m3sconfig = Config(self)
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(
                    error=exception)
            ) from exception

        framework_address = get_framework_address(
            self.get_framework_id(argv), master, config
        )

        data = http.read_endpoint(
            framework_address, "/api/m3s/v0/server/config", self.m3sconfig)

        print(data)

    def version(self, argv):
        """
        Get the version information of Kubernetes
        """

        try:
            master = self.config.master()
            config = self.config
            # pylint: disable=attribute-defined-outside-init
            self.framework_name = argv["<framework-name>"]
            self.m3sconfig = Config(self)            
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(
                    error=exception)
            ) from exception

        framework_address = get_framework_address(
            self.get_framework_id(argv), master, config
        )
        data = http.read_endpoint(
            framework_address, "/api/m3s/v0/server/version", config)

        print(data)

    def status(self, argv):
        """
        Get the status information of these framework and or Kubernetes
        """

        try:
            master = self.config.master()
            config = self.config
            # pylint: disable=attribute-defined-outside-init
            self.framework_name = argv["<framework-name>"]
            self.m3sconfig = Config(self)            
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(
                    error=exception)
            ) from exception

        framework_address = get_framework_address(
            self.get_framework_id(argv), master, config
        )

        if argv["--m3s"]:
            data = http.read_endpoint(
                framework_address, "/api/m3s/v0/status/m3s", config)
            print(data)

        if argv["--kubernetes"]:
            data = http.read_endpoint(
                framework_address, "/api/m3s/v0/status/k8s", config)
            print(data)

    def list(self, argv):
        """
        Show a list of running M3s frameworks
        """

        try:
            master = self.config.master()
            config = self.config
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(
                    error=exception)
            ) from exception

        data = get_frameworks(master, config)
        try:
            table = Table(["ID", "Active", "WebUI", "Name"])
            for framework in data:
                if (
                    not argv["--all"] and framework["active"] is not True
                ) or "m3s" not in framework["name"].lower():
                    continue

                active = "False"
                if framework["active"]:
                    active = "True"

                table.add_row(
                    [framework["id"], active, framework["webui_url"], framework["name"]]
                )
        except Exception as exception:
            raise CLIException(
                "Unable to build table of M3s frameworks: {error}".format(
                    error=exception
                )
            ) from exception

        print(str(table))

    def get_framework_id(self, argv):
        """
        Resolv the id of a framework by the name of a framework
        """

        if argv["<framework-name>"].count("-") != 5:
            data = get_frameworks(self.config.master(), self.config)
            for framework in data:
                if (
                    framework["active"] is not True
                    or framework["name"].lower() != argv["<framework-name>"].lower()
                ):
                    continue
                return framework["id"]
        return argv["<framework-name>"]


    def write_endpoint(self, addr, endpoint, config, method, filename=None):
        """
        Read the specified endpoint and return the results.
        """

        try:
            addr = cli.util.sanitize_address(addr)
        except Exception as exception:
            raise CLIException(
                "Unable to sanitize address '{addr}': {error}".format(
                    addr=addr, error=str(exception)
                )
            )
        try:
            url = "{addr}{endpoint}".format(addr=addr, endpoint=endpoint)
            if config.principal() is not None and config.secret() is not None:
                headers = urllib3.make_headers(
                    basic_auth=config.principal() + ":" + config.secret(),
                )
            else:
                headers = None
            http = urllib3.PoolManager()
            content = ""
            if filename is not None:
                data = open(filename, "rb")
                content = data.read()
            http_response = http.request(
                method,
                url,
                headers=headers,
                body=content,
                timeout=config.agent_timeout(),
            )
            return http_response.data.decode("utf-8")

        except Exception as exception:
            raise CLIException(
                "Unable to open url '{url}': {error}".format(
                    url=url, error=str(exception)
                )
            )
