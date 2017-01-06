package sockets

import (
	"net"
)

type ContainerInfo struct {
	ID                string
	DockerEnvironment []string
	HostName          string
	HostIP            net.IP
}

type SocketInfo struct {
	SrcIP, DstIP     net.IP
	SrcPort, DstPort uint16

	Container *ContainerInfo
	UID       uint32
	Inode     uint64
}
