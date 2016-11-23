// +build !integration

package proc_net_tcp

import (
	"os"
	"testing"
)

func assertIntArraysAreEqual(t *testing.T, expected []int, result []int) bool {
	for _, ex := range expected {
		found := false
		for _, res := range result {
			if ex == res {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected array %v but got %v", expected, result)
			return false
		}
	}
	return true
}

func TestParse_Proc_Net_Tcp(t *testing.T) {
	file, err := os.Open("../../tests/files/proc_net_tcp.txt")
	if err != nil {
		t.Fatalf("Opening ../../tests/files/proc_net_tcp.txt: %s", err)
	}
	socketInfo, err := parseProcNetTCP(file, false)
	if err != nil {
		t.Fatalf("Parse_Proc_Net_Tcp: %s", err)
	}
	if len(socketInfo) != 32 {
		t.Error("expected socket information on 32 sockets but got", len(socketInfo))
	}
	if socketInfo[31].SrcIP.String() != "192.168.2.243" {
		t.Error("Failed to parse source IP address 192.168.2.243")
	}
	if socketInfo[31].SrcPort != 41622 {
		t.Error("Failed to parse source port 41622")
	}
}

func TestParse_Proc_Net_Tcp6(t *testing.T) {
	file, err := os.Open("../../tests/files/proc_net_tcp6.txt")
	if err != nil {
		t.Fatalf("Opening ../../tests/files/proc_net_tcp6.txt: %s", err)
	}
	socketInfo, err := parseProcNetTCP(file, true)
	if err != nil {
		t.Fatalf("Parse_Proc_Net_Tcp: %s", err)
	}
	if len(socketInfo) != 6 {
		t.Error("expected socket information on 6 sockets but got", len(socketInfo))
	}
	if socketInfo[5].SrcIP.String() != "::" {
		t.Error("Failed to parse source IP address ::, got instead", socketInfo[5].SrcIP.String())
	}
	// TODO add an example of a 'real' IPv6 address
	if socketInfo[5].SrcPort != 59497 {
		t.Error("Failed to parse source port 59497, got instead", socketInfo[5].SrcPort)
	}
}
