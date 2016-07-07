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
	if rand.Int() % 2 == 0 {
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
	return &procs.SocketInfo{randIp(), randIp(), port, 0, uint16(rand.Int()), rand.Int63()}
}

func incomingConnection(localPort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{randIp(), randIp(), localPort, uint16(rand.Int()), uint16(rand.Int()), rand.Int63()}
}

func outgoingConnection(remoteIp net.IP, remotePort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{randIp(), remoteIp, uint16(rand.Int()), remotePort, uint16(rand.Int()), rand.Int63()}
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
