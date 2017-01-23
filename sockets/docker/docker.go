package docker

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/fsouza/go-dockerclient"
	"github.com/raboof/connbeat/sockets"
	"github.com/raboof/connbeat/sockets/proc_net_tcp"
)

type Poller struct {
	client             *docker.Client
	environment        map[string]struct{}
	dockerhostIP       net.IP
	dockerhostHostname string
}

func getDockerhostHostname(client *docker.Client) (string, error) {
	if name := os.Getenv("DOCKERHOST_HOSTNAME"); name != "" {
		return name, nil
	}

	info, err := client.Info()
	if err != nil {
		return "", err
	}

	return info.Name, nil
}

func getDockerhostIP(name string) (net.IP, error) {
	if ip := os.Getenv("DOCKERHOST_IP"); ip != "" {
		return net.ParseIP(ip), nil
	}

	ip, err := net.ResolveIPAddr("ip", name)
	if err != nil {
		return nil, err
	}
	return ip.IP, err
}

func New(environment []string) (*Poller, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return new(client, environment)
}

func new(client *docker.Client, environment []string) (*Poller, error) {
	if err := client.Ping(); err != nil {
		return nil, errors.New(fmt.Sprint("Could not connect to docker: ", err))
	}
	env := make(map[string]struct{})
	for _, key := range environment {
		env[key] = struct{}{}
	}

	name, err := getDockerhostHostname(client)
	if err != nil {
		return nil, err
	}

	ip, err := getDockerhostIP(name)
	if err != nil {
		logp.Warn("Could not determine IP address of docker host %s", name)
	}

	return &Poller{
		client:             client,
		environment:        env,
		dockerhostHostname: name,
		dockerhostIP:       ip,
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
	containerInfo, err := p.getContainerInfo(container)
	if err != nil {
		return err
	}
	socks, err := proc_net_tcp.ParseProcNetTCP(&stdout, ipv6, containerInfo)
	if err != nil {
		return err
	}
	for _, s := range socks {
		socketInfo <- s
	}
	return nil
}

func (p *Poller) getContainerInfo(container docker.APIContainers) (*sockets.ContainerInfo, error) {
	inspected, err := p.client.InspectContainer(container.ID)
	if err != nil {
		return nil, err
	}
	environment := p.getEnvironment(inspected)
	return &sockets.ContainerInfo{
		ID:                 container.ID,
		Name:               inspected.Name,
		Image:              container.Image,
		DockerEnvironment:  environment,
		Ports:              inspected.NetworkSettings.Ports,
		DockerhostHostname: p.dockerhostHostname,
		DockerhostIP:       p.dockerhostIP,
	}, nil
}

func (p *Poller) getEnvironment(container *docker.Container) []string {
	result := make([]string, 0, len(p.environment))
	for _, entry := range container.Config.Env {
		key := strings.Split(entry, "=")[0]
		if _, contains := p.environment[key]; contains {
			result = append(result, entry)
		}
	}
	return result
}
