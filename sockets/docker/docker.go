package docker

import (
	"bytes"
	"errors"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/raboof/connbeat/sockets"
	"github.com/raboof/connbeat/sockets/proc_net_tcp"
)

type Poller struct {
	client      *docker.Client
	environment map[string]struct{}
}

func New(environment []string) (*Poller, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	if err = client.Ping(); err != nil {
		return nil, err
	}
	env := make(map[string]struct{})
	for _, key := range environment {
		env[key] = struct{}{}
	}

	return &Poller{
		client:      client,
		environment: env,
	}, nil
}

func (p *Poller) PollCurrentConnections(socketInfo chan<- *sockets.SocketInfo) error {
	containers, err := p.client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		return err
	}
	for _, container := range containers {
		if err = p.pollCurrentConnections(container, socketInfo); err != nil {
			return err
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
		panic(err)
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
	environment, err := p.getEnvironment(container.ID)
	if err != nil {
		return err
	}
	containerInfo := &sockets.ContainerInfo{
		ID:                container.ID,
		DockerEnvironment: environment,
	}
	socks, err := proc_net_tcp.ParseProcNetTCP(&stdout, ipv6, containerInfo)
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

func (p *Poller) getEnvironment(containerId string) ([]string, error) {
	container, err := p.client.InspectContainer(containerId)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(p.environment))
	for _, entry := range container.Config.Env {
		key := strings.Split(entry, "=")[0]
		if _, contains := p.environment[key]; contains {
			result = append(result, entry)
		}
	}
	return result, nil
}
