package tcp_diag

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/packetbeat/procs"
	"github.com/eleme/netlink"
)

func port(bytes [2]byte) uint16 {
	return binary.BigEndian.Uint16([]byte{bytes[0], bytes[1]})
}

func pollConnections(family uint8, socket *netlink.NetlinkSocket, socketInfo chan<- *procs.SocketInfo) error {
	data := netlink.NewInetDiagReqV2(family, syscall.IPPROTO_TCP, netlink.TCP_ALL)
	req := &netlink.NetlinkRequest{
		NlMsghdr: syscall.NlMsghdr{
			Len:   uint32(0),
			Type:  uint16(netlink.SOCK_DIAG_BY_FAMILY),
			Flags: uint16(syscall.NLM_F_DUMP | syscall.NLM_F_REQUEST),
			Seq:   uint32(0),
			Pid:   uint32(os.Getpid()),
		},
	}
	req.AddData(data)
	err := socket.Send(req)
	if err != nil {
		return err
	}
	responses, err := socket.Receive()
	if err != nil {
		fmt.Println("Error receiving netlink results")
		return err
	}
	for _, msg := range responses {
		inetDiagMsg := netlink.ParseInetDiagMsg(msg.Data)
		socketInfo <- &procs.SocketInfo{
			Src_ip:   inetDiagMsg.Id.SrcIP(),
			Dst_ip:   inetDiagMsg.Id.DstIP(),
			Src_port: port(inetDiagMsg.Id.IDiagSPort),
			Dst_port: port(inetDiagMsg.Id.IDiagDPort),

			Uid:   inetDiagMsg.IDiagUid,
			Inode: uint64(inetDiagMsg.IDiagInode),
		}
	}
	return nil
}

func pollCurrentConnectionsForFamily(family uint8, socketInfo chan<- *procs.SocketInfo) error {
	socket, err := netlink.NewNetlinkSocket(netlink.INET_DIAG, 0)
	if err != nil {
		return err
	}
	return pollConnections(family, socket, socketInfo)
}

func pollCurrentConnections(socketInfo chan<- *procs.SocketInfo) error {
	err := pollCurrentConnectionsForFamily(syscall.AF_INET, socketInfo)
	if err != nil {
		return err
	}
	return pollCurrentConnectionsForFamily(syscall.AF_INET6, socketInfo)
}

func GetSocketInfo(pollInterval time.Duration, socketInfo chan<- *procs.SocketInfo) error {
	for {
		// For now we poll periodically
		err := pollCurrentConnections(socketInfo)
		if err != nil {
			logp.Err("Polling connections: %s", err)
		}
		time.Sleep(pollInterval)
	}
}
