package processes

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stvp/assert"
)

func TestCmdline(t *testing.T) {
	procs, err := processes()
	assert.Nil(t, err)
	for _, process := range procs {
		if strings.Contains(process.Cmdline, "processes.test") {
			return
		}
	}
	fmt.Println("No process found with cmdline containing 'processes.test'")
	t.Fail()
}
