// +build cgo

package tcp_diag

import (
	// #cgo LDFLAGS: -l:libmnl.a
	// #include <libmnl/libmnl.h>
	// int poll(int sockfd, const struct mnl_socket * sock);
	"C"

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

func pollCurrentConnections(fd C.int, sock *C.struct_mnl_socket) error {
	_, err := C.poll(fd, sock)
	if err != nil {
		return err
	}

	return nil
}

func GetSocketInfo(pollInterval time.Duration, socketInfo chan<- *procs.SocketInfo) error {
	socketInfoChan = socketInfo

	sock, err := C.mnl_socket_open(C.NETLINK_INET_DIAG)
	if err != nil {
		return err
	}
	fd, err := C.mnl_socket_get_fd(sock)
	if err != nil {
		return err
	}

	for {
		err := pollCurrentConnections(fd, sock)
		if err != nil {
			C.mnl_socket_close(sock)
			return err
		}
		time.Sleep(pollInterval)
	}
}
