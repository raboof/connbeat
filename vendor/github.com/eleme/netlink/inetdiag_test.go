package netlink

import (
	"testing"
	"unsafe"
)

func TestSizeofInetDiagReqV2(t *testing.T) {
	req := InetDiagReqV2{}
	if unsafe.Sizeof(req) != SizeofInetDiagReqV2 {
		t.Error("size of InetDiagReqV2 error")
	}
}
