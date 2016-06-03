// +build linux

// Originally inspired by the MIT-licensed
// https://github.com/mitchellh/go-ps/blob/master/process_unix.go

package processes

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/elastic/beats/packetbeat/procs"
)

type Processes struct {
	byInode       map[int64]*UnixProcess
	exposeCmdline bool
	exposeEnviron bool
}

func New(exposeCmdline, exposeEnviron bool) *Processes {
	return &Processes{
		byInode:       make(map[int64]*UnixProcess),
		exposeCmdline: exposeCmdline,
		exposeEnviron: exposeEnviron,
	}
}

func (ps *Processes) Refresh() error {
	procs, err := processes(ps.exposeCmdline, ps.exposeEnviron)
	if err != nil {
		return err
	}
	for _, p := range procs {
		err := p.Refresh(ps.exposeCmdline, ps.exposeEnviron)
		if err != nil {
			return err
		}
		for _, inode := range p.inodes {
			ps.byInode[inode] = p
		}
	}

	return nil
}

// UnixProcess is an implementation of Process that contains Unix-specific
// fields and information.
type UnixProcess struct {
	pid    int
	ppid   int
	state  rune
	pgrp   int
	sid    int
	inodes []int64

	Binary  string
	Cmdline string
	Environ string
}

func (p *UnixProcess) Pid() int {
	return p.pid
}

func (p *UnixProcess) PPid() int {
	return p.ppid
}

// Refresh reloads all the data associated with this process.
func (p *UnixProcess) Refresh(exposeCmdline, exposeEnviron bool) error {
	prefix := ""
	data, err := readFile(prefix, p.pid, "stat")
	if err != nil {
		return err
	}

	// First, parse out the image name
	binStart := strings.IndexRune(data, '(') + 1
	binEnd := strings.IndexRune(data[binStart:], ')')
	p.Binary = data[binStart : binStart+binEnd]

	// Move past the image name and start parsing the rest
	data = data[binStart+binEnd+2:]
	_, err = fmt.Sscanf(data,
		"%c %d %d %d",
		&p.state,
		&p.ppid,
		&p.pgrp,
		&p.sid)

	if err != nil {
		return err
	}

	inodes, err := procs.FindSocketsOfPid(prefix, p.pid)
	p.inodes = inodes

	if exposeCmdline {
		p.Cmdline, err = readFile(prefix, p.pid, "cmdline")
	}
	if exposeEnviron {
		p.Environ, err = readFile(prefix, p.pid, "environ")
	}

	return err
}

func readFile(prefix string, pid int, filename string) (string, error) {
	path := fmt.Sprintf("%s/proc/%d/%s", prefix, pid, filename)
	bytes, err := ioutil.ReadFile(path)
	return string(bytes), err
}

func processes(exposeCmdline, exposeEnviron bool) ([]*UnixProcess, error) {
	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	results := make([]*UnixProcess, 0, 50)
	for {
		fis, err := d.Readdir(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, fi := range fis {
			// We only care about directories, since all pids are dirs
			if !fi.IsDir() {
				continue
			}

			// We only care if the name starts with a numeric
			name := fi.Name()
			if name[0] < '0' || name[0] > '9' {
				continue
			}

			// From this point forward, any errors we just ignore, because
			// it might simply be that the process doesn't exist anymore.
			pid, err := strconv.ParseInt(name, 10, 0)
			if err != nil {
				continue
			}

			p, err := newUnixProcess(int(pid), exposeCmdline, exposeEnviron)
			if err != nil {
				continue
			}

			results = append(results, p)
		}
	}

	return results, nil
}

func newUnixProcess(pid int, exposeCmdline, exposeEnviron bool) (*UnixProcess, error) {
	p := &UnixProcess{pid: pid}
	return p, p.Refresh(exposeCmdline, exposeEnviron)
}

func (ps *Processes) FindProcessByInode(inode int64) *UnixProcess {
	proc := ps.byInode[inode]
	if proc == nil {
		// Refesh and try again
		ps.Refresh()

		proc = ps.byInode[inode]
		if proc == nil {
			return &UnixProcess{
				Binary: fmt.Sprintf("Unknown process with inode %d", inode),
			}
		}
		return proc
	}
	return proc
}
