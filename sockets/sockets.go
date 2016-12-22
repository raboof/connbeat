package sockets

import (
	"net"
)

type SocketInfo struct {
	SrcIP, DstIP     net.IP
	SrcPort, DstPort uint16

	ContainerId string
	UID         uint32
	Inode       uint64
}
