from __future__ import print_function
import sys

import os
import stat
import connbeat
from nose.plugins.attrib import attr

def eprint(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)

class DockerTest(connbeat.BaseTest):
    def should_contain(self, output, check, error):
        for evt in output:
            if check(evt):
                return
        self.assertFalse(error)

    @attr('integration')
    def test_docker_connection(self):
        """
        Basic connections from peer docker containers are published
        """
        self.render_config_template(
            enable_local_connections = False,
            enable_docker = True
        )

        proc = self.start_beat()
        self.wait_until(lambda: self.output_lines() > 1, max_timeout = 20)
        proc.check_kill_and_wait()

        output = self.read_output_json()

        for line in output:
            eprint(line)

        # docker-compose.yml specifies an nginx peer container should be started
        self.should_contain(output, lambda e: e['local_port'] == 80, "process listening on port 80")
