package docker

import (
	"bytes"
	"errors"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/fsouza/go-dockerclient"
	"github.com/raboof/connbeat/sockets"
	"github.com/raboof/connbeat/sockets/proc_net_tcp"
)

type Poller struct {
	client *docker.Client
}

func New() (*Poller, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	if err = client.Ping(); err != nil {
		return nil, err
	}

	return &Poller{
		client: client,
	}, nil
}

func (p *Poller) PollCurrentConnections(socketInfo chan<- *sockets.SocketInfo) error {
	containers, err := p.client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		return err
	}
	for _, container := range containers {
		if err = p.pollCurrentConnections(container, socketInfo); err != nil {
			logp.Warn("Failed to poll connections for container %s (%s): %s", container.ID, container.Image, err)
		}
	}
	return nil
}

func (p *Poller) pollCurrentConnections(container docker.APIContainers, socketInfo chan<- *sockets.SocketInfo) error {
	err := p.pollCurrentConnectionsFor(container, "/proc/net/tcp", false, socketInfo)
	if err != nil {
		return err
	}
	return p.pollCurrentConnectionsFor(container, "/proc/net/tcp6", true, socketInfo)
}

func (p *Poller) pollCurrentConnectionsFor(container docker.APIContainers, file string, ipv6 bool, socketInfo chan<- *sockets.SocketInfo) error {
	exec, err := p.client.CreateExec(docker.CreateExecOptions{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"cat", file},
		Container:    container.ID,
		Privileged:   false,
	})
	if err != nil {
		return err
	}
	var stdout, stderr bytes.Buffer
	if err = p.client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: &stdout,
		ErrorStream:  &stderr,
		RawTerminal:  false,
	}); err != nil {
		return err
	}
	result, err := p.client.InspectExec(exec.ID)
	if err != nil {
		return err
	}
	if result.Running {
		return errors.New("exec was still running?")
	}
	if result.ExitCode != 0 {
		return errors.New("exit code was not 0")
	}
	socks, err := proc_net_tcp.ParseProcNetTCP(&stdout, ipv6, container.ID)
	if err != nil {
		return err
	}
	for _, s := range socks {
		if s.Inode != 0 {
			socketInfo <- s
		}
	}
	return nil
}
