package tcp_diag

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

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
		logp.Err("Error receiving netlink results: %s", err)
		return err
	}
	for _, msg := range responses {
		if msg.Header.Type == syscall.NLMSG_ERROR {
			msgerr := (*syscall.NlMsgerr)(unsafe.Pointer(&msg.Data[0]))
			return errors.New(fmt.Sprintf("Netlink returned error message with error code %d: %s",
				-msgerr.Error,
				syscall.Errno(-msgerr.Error).Error()))
		} else {
			inetDiagMsg := netlink.ParseInetDiagMsg(msg.Data)
			fmt.Printf("Processing netlink response for remote port %d\n", inetDiagMsg.Id.IDiagDPort)
			socketInfo <- &procs.SocketInfo{
				SrcIP:   inetDiagMsg.Id.SrcIP(),
				DstIP:   inetDiagMsg.Id.DstIP(),
				SrcPort: port(inetDiagMsg.Id.IDiagSPort),
				DstPort: port(inetDiagMsg.Id.IDiagDPort),

				UID:   inetDiagMsg.IDiagUid,
				Inode: uint64(inetDiagMsg.IDiagInode),
			}
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
			logp.Err("Polling connections: %s, degrading to /proc/net/tcp-based monitoring", err)
			return err
		}
		time.Sleep(pollInterval)
	}
}
