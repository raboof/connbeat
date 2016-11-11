package beater

import (
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/elastic/beats/packetbeat/procs"
	"github.com/stvp/assert"
)

func randByte() byte {
	return byte(rand.Intn(256))
}

func randIP() net.IP {
	if rand.Int()%2 == 0 {
		return net.IPv4(randByte(), randByte(), randByte(), randByte())
	} else {
		ip := make(net.IP, net.IPv6len)
		for i := 0; i < 16; i++ {
			ip[i] = randByte()
		}
		return ip
	}
}

func listeningConnection(port uint16) *procs.SocketInfo {
	return &procs.SocketInfo{
		SrcIP:   randIP(),
		DstIP:   randIP(),
		SrcPort: port,
		DstPort: 0,
		UID:     uint32(rand.Int()),
		Inode:   uint64(rand.Int()),
	}
}

func incomingConnection(localPort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{
		SrcIP:   randIP(),
		DstIP:   randIP(),
		SrcPort: localPort,
		DstPort: uint16(rand.Int()),
		UID:     uint32(rand.Int()),
		Inode:   uint64(rand.Int63()),
	}
}

func outgoingConnection(remoteIP net.IP, remotePort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{
		SrcIP:   randIP(),
		DstIP:   remoteIP,
		SrcPort: uint16(rand.Int()),
		DstPort: remotePort,
		UID:     uint32(rand.Int()),
		Inode:   uint64(rand.Int63()),
	}
}

func TestDeduplicateListeningSockets(t *testing.T) {
	input := make(chan *procs.SocketInfo, 0)
	connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

	go filterAndPublish(true, true, true, 5*time.Second, input, connections, servers)

	input <- listeningConnection(80)
	_, ok := <-servers
	assert.Equal(t, ok, true, "a server should be reported")
	input <- listeningConnection(80)

	time.Sleep(100 * time.Millisecond)

	select {
	case <-connections:
		t.Fail()
	case <-servers:
		t.Fail()
	default:
		// Nothing to read: OK!
	}
}

func TestFilterConnectionsAssociatedWithListeningSockets(t *testing.T) {
	input := make(chan *procs.SocketInfo, 0)
	connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

	go filterAndPublish(true, true, true, 5*time.Second, input, connections, servers)

	input <- listeningConnection(80)
	_, ok := <-servers
	assert.Equal(t, ok, true, "a server should be reported")
	input <- incomingConnection(80)

	time.Sleep(100 * time.Millisecond)

	select {
	case <-connections:
		t.Fail()
	case <-servers:
		t.Fail()
	default:
		// Nothing to read: OK!
	}
}

func TestDedupClientConnections(t *testing.T) {
	input := make(chan *procs.SocketInfo, 0)
	connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

	go filterAndPublish(true, true, true, 5*time.Second, input, connections, servers)

	remoteIp := randIP()
	input <- outgoingConnection(remoteIp, 80)
	_, ok := <-connections
	assert.Equal(t, ok, true, "a client connection should be reported")
	input <- outgoingConnection(remoteIp, 80)

	time.Sleep(100 * time.Millisecond)

	select {
	case <-connections:
		t.Fail()
	case <-servers:
		t.Fail()
	default:
		// Nothing to read: OK!
	}
}

func TestRepublishOldClientConnections(t *testing.T) {
	input := make(chan *procs.SocketInfo, 0)
	connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

	go filterAndPublish(false, false, true, 0*time.Second, input, connections, servers)

	remoteIp := randIP()
	input <- outgoingConnection(remoteIp, 80)
	_, ok := <-connections
	assert.Equal(t, ok, true, "a client connection should be reported")
	input <- outgoingConnection(remoteIp, 80)
	_, ok = <-connections
	assert.Equal(t, ok, true, "another client connection should be reported")
}
