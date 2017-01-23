from __future__ import print_function
import sys

import os
import stat
import connbeat
from nose.plugins.attrib import attr

def eprint(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)

class ConnectionTest(connbeat.BaseTest):
    def should_contain(self, output, check, error):
        for evt in output:
            if check(evt):
                return
        self.assertFalse(error)

    @attr('integration')
    def test_connection(self):
        """
        Basic connections are published
        """
        self.render_config_template()
        os.environ['PROC_NET_TCP'] = '../../tests/files/proc-net-tcp-test-small'
        os.environ['PROC_NET_TCP6'] = '../../tests/files/proc-net-tcp6-test-empty'

        proc = self.start_beat()
        self.wait_until(lambda: self.output_lines() > 0)
        proc.check_kill_and_wait()

        output = self.read_output_json()

        for line in output:
            eprint(line)

        self.should_contain(output, lambda e: e['local_port'] == 80, "process listening on port 80")

        self.should_contain(output, lambda e: e['local_port'] == 631, "process listening on port 631")

        self.should_contain(output, lambda e: e['local_port'] == 40074, "process listening on port 40074")

        self.should_contain(
            output,
            lambda e: e['beat']['local_ips'] == ['192.168.2.243'],
            "record 192.168.2.243 as local IP")
