package processes

import (
	"bytes"
	"os/exec"
	"strings"
	"strconv"

	"github.com/elastic/beats/libbeat/logp"
)

// let's make this an interface later:
type UnixProcess struct {
	pid    int
	inodes []int64

	Binary  string
	Cmdline string
	Environ string
}

type Processes struct {
	byPid         map[int64]*UnixProcess
	exposeCmdline bool
	exposeEnviron bool
}

func New(exposeCmdline, exposeEnviron bool) *Processes {
	return &Processes{
		byPid:         make(map[int64]*UnixProcess),
		exposeCmdline: exposeCmdline,
		exposeEnviron: exposeEnviron,
	}
}

func (ps *Processes) FindProcessByInode(inode int64) *UnixProcess {
	return nil
}

func (ps *Processes) services() map[int64]string {
	result := make(map[int64]*UnixProcess)

	cmd := exec.Command("tasklist", "/svc")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logp.Err("opening stdout: %s", err)
	}
	err = cmd.Start()
	if err != nil {
		logp.Err("starting tasklist: %s", err)
	}
	buf := new(bytes.Buffer)
	// TODO eventually we'd want to stream this, for now just get the whole thing:
	buf.ReadFrom(stdout)
	lines := strings.Split(buf.String(), "\n")
	pidIdx = strings.Index(lines[1], " ") + 1
	servicesIdx = strings.Index(lines[1], " ") + 1
	for _, line := range lines[2:] {
		pid, err := strconv.Atoi(strings.TrimSpace(line[pidIdx:servicesIdx]))
		services := string.TrimSpace(line[servicesIdx:])
		result[pid] =
	}

	return result, nil
}

func (ps *Processes) refreshProcesses() {
  cmd := exec.Command("wmic", "path", "win32_process", "get", "Processid,Caption,Commandline")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logp.Err("opening stdout: %s", err)
	}
	err = cmd.Start()
	if err != nil {
		logp.Err("starting wmic: %s", err)
	}
	buf := new(bytes.Buffer)
	// TODO eventually we'd want to stream this, for now just get the whole thing:
	buf.ReadFrom(stdout)
	lines := strings.Split(buf.String(), "\n")
	header := lines[0]
	cmdlineIdx := strings.Index(header, "CommandLine")
	pidIdx := strings.Index(header, "ProcessId")

	for _, line := range lines[1:] {
		if (len(line) > pidIdx) {
			caption := strings.TrimSpace(line[0:cmdlineIdx])
			cmdline := strings.TrimSpace(line[cmdlineIdx:pidIdx])
			pid, _ := strconv.ParseInt(strings.TrimSpace(line[pidIdx:]), 10, 64)
			//fmt.Printf("One proc: %s, %s, %s", caption, cmdline, pidx)
			// fmt.Printf("One proc: %s, %s, %s\n", caption, cmdline, pid)

			ps.byPid[pid] = &UnixProcess{
				Binary: caption,
				Cmdline: cmdline,
				pid: int(pid),
			}
		}
	}
}

func (ps *Processes) FindProcessByPid(pid int64) *UnixProcess {
	proc := ps.byPid[pid]
	if proc == nil {
		ps.refreshProcesses()
		return ps.byPid[pid]
	}
	return proc
}
