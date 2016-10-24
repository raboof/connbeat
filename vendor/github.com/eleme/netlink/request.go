package netlink

import (
	"os"
	"syscall"
	"unsafe"
)

type NetlinkRequestData interface {
	Len() int
	Serialize() []byte
}

// linux/netlink.h
type NetlinkRequest struct {
	syscall.NlMsghdr
	Data []NetlinkRequestData
}

func (req *NetlinkRequest) Serialize() []byte {
	length := syscall.SizeofNlMsghdr
	dataBytes := make([][]byte, len(req.Data))
	for i, data := range req.Data {
		dataBytes[i] = data.Serialize()
		length = length + len(dataBytes[i])
	}
	req.Len = uint32(length)
	b := make([]byte, length)
	hdr := (*(*[syscall.SizeofNlMsghdr]byte)(unsafe.Pointer(req)))[:]
	next := syscall.SizeofNlMsghdr
	copy(b[0:next], hdr)
	for _, data := range dataBytes {
		for _, dataByte := range data {
			b[next] = dataByte
			next = next + 1
		}
	}
	return b
}

func (req *NetlinkRequest) AddData(data NetlinkRequestData) {
	if data != nil {
		req.Data = append(req.Data, data)
	}
}

func NewNetlinkRequest() *NetlinkRequest {
	return &NetlinkRequest{
		NlMsghdr: syscall.NlMsghdr{
			Len:   uint32(0),
			Type:  uint16(syscall.NLMSG_DONE),
			Flags: uint16(0),
			Seq:   uint32(0),
			Pid:   uint32(os.Getpid()),
		},
	}
}
