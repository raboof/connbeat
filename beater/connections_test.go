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

func randIp() net.IP {
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
		Src_ip:   randIp(),
		Dst_ip:   randIp(),
		Src_port: port,
		Dst_port: 0,
		Uid:      uint16(rand.Int()),
		Inode:    rand.Int63(),
	}
}

func incomingConnection(localPort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{
		Src_ip:   randIp(),
		Dst_ip:   randIp(),
		Src_port: localPort,
		Dst_port: uint16(rand.Int()),
		Uid:      uint16(rand.Int()),
		Inode:    rand.Int63(),
	}
}

func outgoingConnection(remoteIp net.IP, remotePort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{
		Src_ip:   randIp(),
		Dst_ip:   remoteIp,
		Src_port: uint16(rand.Int()),
		Dst_port: remotePort,
		Uid:      uint16(rand.Int()),
		Inode:    rand.Int63(),
	}
}

func TestDeduplicateListeningSockets(t *testing.T) {
	input := make(chan *procs.SocketInfo, 0)
	connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

	go filterAndPublish(true, true, 5*time.Second, input, connections, servers)

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

	go filterAndPublish(true, true, 5*time.Second, input, connections, servers)

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

	go filterAndPublish(true, true, 5*time.Second, input, connections, servers)

	remoteIp := randIp()
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

	go filterAndPublish(false, false, 0*time.Second, input, connections, servers)

	remoteIp := randIp()
	input <- outgoingConnection(remoteIp, 80)
	_, ok := <-connections
	assert.Equal(t, ok, true, "a client connection should be reported")
	input <- outgoingConnection(remoteIp, 80)
	_, ok = <-connections
	assert.Equal(t, ok, true, "another client connection should be reported")
}
