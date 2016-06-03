package beater

import (
	"github.com/stvp/assert"
	"math/rand"
	"testing"
	"time"

	"github.com/elastic/beats/packetbeat/procs"
)

func listeningConnection(port uint16) *procs.SocketInfo {
	return &procs.SocketInfo{rand.Uint32(), rand.Uint32(), port, 0, uint16(rand.Int()), rand.Int63()}
}

func incomingConnection(localPort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{rand.Uint32(), rand.Uint32(), localPort, uint16(rand.Int()), uint16(rand.Int()), rand.Int63()}
}

func outgoingConnection(remoteIp uint32, remotePort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{rand.Uint32(), remoteIp, uint16(rand.Int()), remotePort, uint16(rand.Int()), rand.Int63()}
}

func TestDeduplicateListeningSockets(t *testing.T) {
	input := make(chan *procs.SocketInfo, 0)
	connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

	go filterAndPublish(5*time.Second, input, connections, servers)

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

	go filterAndPublish(5*time.Second, input, connections, servers)

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

	go filterAndPublish(5*time.Second, input, connections, servers)

	input <- outgoingConnection(6543142, 80)
	_, ok := <-connections
	assert.Equal(t, ok, true, "a client connection should be reported")
	input <- outgoingConnection(6543142, 80)

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

	go filterAndPublish(0*time.Second, input, connections, servers)

	input <- outgoingConnection(6543142, 80)
	_, ok := <-connections
	assert.Equal(t, ok, true, "a client connection should be reported")
	input <- outgoingConnection(6543142, 80)
	_, ok = <-connections
	assert.Equal(t, ok, true, "another client connection should be reported")
}
