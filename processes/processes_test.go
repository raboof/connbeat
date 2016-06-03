package processes

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stvp/assert"
)

func TestCmdline(t *testing.T) {
	procs, err := processes(true, false)
	assert.Nil(t, err)
	for _, process := range procs {
		if strings.Contains(process.Cmdline, "processes.test") {
			return
		}
	}
	fmt.Println("No process found with cmdline containing 'processes.test'")
	t.Fail()
}

func TestNoCmdline(t *testing.T) {
	procs, err := processes(false, true)
	assert.Nil(t, err)
	for _, process := range procs {
		if process.Cmdline != "" {
			fmt.Println("Process found with cmdline even though we're configured not to expose that")
			t.Fail()
		}
	}
}
