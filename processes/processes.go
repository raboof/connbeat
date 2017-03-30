// Originally inspired by the MIT-licensed
// https://github.com/mitchellh/go-ps/blob/master/process_unix.go

package processes

type Processes struct {
	byInode       map[uint64]*UnixProcess
	exposeCmdline bool
	exposeEnviron bool
}

func New(exposeCmdline, exposeEnviron bool) *Processes {
	return &Processes{
		byInode:       make(map[uint64]*UnixProcess),
		exposeCmdline: exposeCmdline,
		exposeEnviron: exposeEnviron,
	}
}

// UnixProcess is an implementation of Process that contains Unix-specific
// fields and information.
type UnixProcess struct {
	pid    int
	inodes []uint64

	Binary  string
	Cmdline string
	Environ string
}

func (p *UnixProcess) Pid() int {
	return p.pid
}
