// +build !linux

package processes

import (
	"errors"
)

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

func (ps *Processes) Refresh() error {
	// Only supported under linux
	return nil
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

// Refresh reloads all the data associated with this process.
func (p *UnixProcess) Refresh(exposeCmdline, exposeEnviron bool) error {
	return errors.New("Only supported under linux")
}

func (ps *Processes) FindProcessByInode(inode uint64) *UnixProcess {
	// Only supported under linux
	return nil
}
