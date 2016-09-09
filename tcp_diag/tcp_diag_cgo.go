// +build cgo

package tcp_diag

import (
	//"C"

	"encoding/binary"
	"net"
	"time"

	"github.com/elastic/beats/packetbeat/procs"
)

var socketInfoChan chan<- *procs.SocketInfo

func ipv4ip(original uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, original)
	return ip
}

//export SocketInfoCallback
func SocketInfoCallback(uid uint16, inode int64, src uint32, dst uint32, sport uint16, dport uint16) {
	socketInfoChan <- &procs.SocketInfo{
		Src_ip:   ipv4ip(src),
		Dst_ip:   ipv4ip(dst),
		Src_port: sport,
		Dst_port: dport,

		Uid:   uid,
		Inode: inode,
	}
}

func GetSocketInfo(pollInterval time.Duration, socketInfo chan<- *procs.SocketInfo) error {
	return nil
}
