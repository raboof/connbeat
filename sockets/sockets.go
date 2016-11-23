package sockets

import (
	"net"
)

type SocketInfo struct {
	SrcIP, DstIP     net.IP
	SrcPort, DstPort uint16

	UID   uint32
	Inode uint64
}
