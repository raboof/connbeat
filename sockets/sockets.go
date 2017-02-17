package sockets

import (
	"net"

	"github.com/fsouza/go-dockerclient"
)

type ContainerInfo struct {
	ID                 string
	Name               string
	Image              string
	DockerEnvironment  []string
	DockerLabels	   map[string]string
	Ports              map[docker.Port][]docker.PortBinding
	DockerhostHostname string
	DockerhostIP       net.IP
}

type SocketInfo struct {
	SrcIP, DstIP     net.IP
	SrcPort, DstPort uint16

	Container *ContainerInfo
	UID       uint32
	Inode     uint64
}
