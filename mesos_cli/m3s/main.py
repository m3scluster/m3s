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

from cli.exceptions import CLIException
from cli.plugins import PluginBase
from cli.util import Table

from cli.mesos import get_frameworks, get_framework_address
from cli import http
from cli.exceptions import CLIException


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
            "arguments": ['<framework-id>'],
            "flags": {},
            "short_help": "Get kubernetes configuration file",
            "long_help": "Get kubernetes configuration file"
        },
        "list": {
            "arguments": [],
            "flags": {
                "-a --all": "list all M3s frameworks, not only running [default: False]"
            },
            "short_help": "Show list of running M3s frameworks",
            "long_help": "Show list of running M3s frameworks",
        }
    }

    def kubeconfig(self, argv):
        """
        Load the kubeconfig file from the kubernetes api server
        """

        try:
            master = self.config.master()
            config = self.config
        except Exception as exception:
            raise CLIException("Unable to get leading master address: {error}"
                               .format(error=exception))

        framework_address = get_framework_address(argv["<framework-id>"], master, config)
        data = http.read_endpoint(framework_address, "/v0/server/config", config)

        print (data)


    def list(self, argv):
        """
        Show a list of running M3s frameworks
        """

        try:
            master = self.config.master()
            config = self.config
        except Exception as exception:
            raise CLIException("Unable to get leading master address: {error}"
                               .format(error=exception))

        data = get_frameworks(master, config)
        try:
            table = Table(["ID", "Active", "WebUI", "Name"])
            for framework in data:
                if (not argv["--all"] and framework["active"] is not True) or "m3s" not in framework["name"].lower():
                    continue

                active = "False"
                if framework["active"]:
                    active = "True"

                table.add_row([framework["id"],
                               active,
                               framework["webui_url"],
                               framework["name"]])
        except Exception as exception:
            raise CLIException("Unable to build table of M3s frameworks: {error}"
                               .format(error=exception))

        print(str(table))
        



