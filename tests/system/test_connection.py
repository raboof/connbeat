from __future__ import print_function
import sys

import os
import stat
import connbeat
from nose.plugins.attrib import attr

def eprint(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)

class ConnectionTest(connbeat.BaseTest):
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

        evt = output[0]
        eprint("Event 0")
        eprint(evt)

        self.assertEqual(evt['local_port'], 80)

        evt = output[1]
        eprint("Event 1")
        eprint(evt)
        self.assertEqual(evt['local_port'], 631)

        evt = output[2]
        eprint("Event 2")
        eprint(evt)
        self.assertEqual(evt['local_port'], 40074, "msg here")
        self.assertItemsEqual(evt['beat']['local_ips'], ['192.168.2.243'])
