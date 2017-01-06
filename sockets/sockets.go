package sockets

import (
	"net"

	"github.com/fsouza/go-dockerclient"
)

type ContainerInfo struct {
	ID                string
	DockerEnvironment []string
	HostName          string
	Ports             map[docker.Port][]docker.PortBinding
}

type SocketInfo struct {
	SrcIP, DstIP     net.IP
	SrcPort, DstPort uint16

	Container *ContainerInfo
	UID       uint32
	Inode     uint64
}
