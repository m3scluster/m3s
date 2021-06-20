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
from cli.mesos import get_tasks
from cli.plugins import PluginBase
from cli.util import Table

from cli.mesos import TaskIO

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
            "arguments": ['<framwork-id>'],
            "flags": {},
            "short_help": "Get kubernetes configuration file",
            "long_help": "Get kubernetes configuration file"
        }
    }


    def kubeconfig(self, argv):
        """
        Load the kubeconfig file from the kubernetes api server
        """

