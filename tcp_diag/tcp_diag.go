// +build linux

package tcp_diag

import (
	// #cgo LDFLAGS: -lmnl
	// #include <libmnl/libmnl.h>
	// int poll(int sockfd, const struct mnl_socket * sock);
	"C"

	"time"

	"github.com/elastic/beats/packetbeat/procs"
)

var socketInfoChan chan<- *procs.SocketInfo

//export SocketInfoCallback
func SocketInfoCallback(uid uint16, inode int64, src uint32, dst uint32, sport uint16, dport uint16) {
	socketInfoChan <- &procs.SocketInfo{
		Src_ip:   src,
		Dst_ip:   dst,
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
