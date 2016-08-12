import re
import sys
import unittest
from connbeat import BaseTest


class Test(BaseTest):
    @unittest.skipUnless(re.match("(?i)win|linux|darwin|openbsd", sys.platform), "os")
    def test_start_stop(self):
        """
        Connbeat starts and stops without error.
        """
        self.render_config_template()
        proc = self.start_beat()
        self.wait_until(lambda: self.log_contains("start running"))
        proc.check_kill_and_wait()

        # Ensure no errors or warnings exist in the log.
        log = self.get_log()
        self.assertNotRegexpMatches(log, "ERR|WARN")

        # Ensure all Beater stages are used.
        self.assertRegexpMatches(log, re.compile("(?i).*".join([
            "Setup Beat: connbeat",
            "connbeat start running",
            "Received sigterm/sigint, stopping",
            "connbeat stopped"
        ]), re.DOTALL))
