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

from cli.exceptions import CLIException
from cli.plugins import PluginBase
from cli.util import Table

from cli.mesos import get_frameworks, get_framework_address
from cli import http


PLUGIN_NAME = "m3s"
PLUGIN_CLASS = "M3s"

VERSION = "0.1.0"

SHORT_HELP = "Interacts with the Kubernetes Framework M3s"


class M3s(PluginBase):
    """
    The m3s plugin.
    """

    COMMANDS = {
        "kubeconfig": {
            "arguments": ["<framework-id>"],
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
            "arguments": ["<framework-id>"],
            "flags": {},
            "short_help": "Get the version number of Kubernetes",
            "long_help": "Get the version number of Kubernetes",
        },
        "status": {
            "arguments": ["<framework-id>"],
            "flags": {
                "-m --m3s": "Give out the M3s status.",
                "-k --kubernetes": "Give out the Kubernetes status.",
            },
            "short_help": "Get out live status information",
            "long_help": "Get out live status information",
        },
        "scale": {
            "arguments": ["<framework-id>", "<count>"],
            "flags": {
                "-a --agent": "Scale up/down Kubernetes agents",
                "-e --etcd": "Scale up/down etcd",
            },
            "short_help": "Scale up/down the Manager or Agent of Kubernetes",
            "long_help": "Scale up/down the Manager or Agent of Kubernetes",
        },
    }

    def scale(self, argv):
        """
        Scale up/down Kubernetes Manager or Agents
        """

        try:
            master = self.config.master()
            config = self.config
            # pylint: disable=attribute-defined-outside-init
            self.m3sconfig = self._get_config()
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(error=exception)
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
            self.m3sconfig = self._get_config()
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(error=exception)
            ) from exception

        framework_address = get_framework_address(
            self.get_framework_id(argv), master, config
        )
        data = http.read_endpoint(framework_address, "/api/m3s/v0/server/config", self)

        print(data)

    def version(self, argv):
        """
        Get the version information of Kubernetes
        """

        try:
            master = self.config.master()
            config = self.config
            # pylint: disable=attribute-defined-outside-init
            self.m3sconfig = self._get_config()
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(error=exception)
            ) from exception

        framework_address = get_framework_address(
            self.get_framework_id(argv), master, config
        )
        data = http.read_endpoint(framework_address, "/api/m3s/v0/server/version", self)

        print(data)

    def status(self, argv):
        """
        Get the status information of these framework and or Kubernetes
        """

        try:
            master = self.config.master()
            config = self.config
            # pylint: disable=attribute-defined-outside-init
            self.m3sconfig = self._get_config()
        except Exception as exception:
            raise CLIException(
                "Unable to get leading master address: {error}".format(error=exception)
            ) from exception

        framework_address = get_framework_address(
            self.get_framework_id(argv), master, config
        )

        if argv["--m3s"]:
            data = http.read_endpoint(framework_address, "/api/m3s/v0/status/m3s", self)
            print(data)

        if argv["--kubernetes"]:
            data = http.read_endpoint(framework_address, "/api/m3s/v0/status/k8s", self)
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
                "Unable to get leading master address: {error}".format(error=exception)
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

        if argv["<framework-id>"].count("-") != 5:
            data = get_frameworks(self.config.master(), self.config)
            for framework in data:
                if (
                    framework["active"] is not True
                    or framework["name"].lower() != argv["<framework-id>"].lower()
                ):
                    continue
                return framework["id"]
        return argv["<framework-id>"]

    def principal(self):
        """
        Return the principal in the configuration file
        """

        return self.m3sconfig["m3s"].get("principal")

    def secret(self):
        """
        Return the secret in the configuration file
        """

        return self.m3sconfig["m3s"].get("secret")

    # pylint: disable=no-self-use
    def agent_timeout(self, default=5):
        """
        Return the connection timeout of the agent
        """

        return default

    def _get_config(self):
        """
        Get authentication header for the framework
        """

        try:
            data = toml.load(self.config.path)
        except Exception as exception:
            raise CLIException(
                "Error loading config file as TOML: {error}".format(error=exception)
            ) from exception

        return data
