package docker

import (
	"bytes"
	"errors"

	"github.com/fsouza/go-dockerclient"
	"github.com/raboof/connbeat/sockets"
	"github.com/raboof/connbeat/sockets/proc_net_tcp"
)

func PollCurrentConnections(endpoint string, socketInfo chan<- *sockets.SocketInfo) error {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return err
	}
	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		return err
	}
	for _, container := range containers {
		if err = pollCurrentConnections(client, container, socketInfo); err != nil {
			return err
		}
	}
	return nil
}

func pollCurrentConnections(client *docker.Client, container docker.APIContainers, socketInfo chan<- *sockets.SocketInfo) error {
	err := pollCurrentConnectionsFor(client, container, "/proc/net/tcp", false, socketInfo)
	if err != nil {
		return err
	}
	return pollCurrentConnectionsFor(client, container, "/proc/net/tcp6", true, socketInfo)
}

func pollCurrentConnectionsFor(client *docker.Client, container docker.APIContainers, file string, ipv6 bool, socketInfo chan<- *sockets.SocketInfo) error {
	exec, err := client.CreateExec(docker.CreateExecOptions{
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
	if err = client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: &stdout,
		ErrorStream:  &stderr,
		RawTerminal:  false,
	}); err != nil {
		return err
	}
	result, err := client.InspectExec(exec.ID)
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
