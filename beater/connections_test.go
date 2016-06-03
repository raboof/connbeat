package beater

import (
	"github.com/stvp/assert"
	"math/rand"
	"testing"
	"time"

	"github.com/elastic/beats/packetbeat/procs"
)

func listeningConnection(port uint16) *procs.SocketInfo {
	return &procs.SocketInfo{
		Src_ip:   rand.Uint32(),
		Dst_ip:   rand.Uint32(),
		Src_port: port,
		Dst_port: 0,
		Uid:      uint16(rand.Int()),
		Inode:    rand.Int63(),
	}
}

func incomingConnection(localPort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{
		Src_ip:   rand.Uint32(),
		Dst_ip:   rand.Uint32(),
		Src_port: localPort,
		Dst_port: uint16(rand.Int()),
		Uid:      uint16(rand.Int()),
		Inode:    rand.Int63(),
	}
}

func outgoingConnection(remoteIp uint32, remotePort uint16) *procs.SocketInfo {
	return &procs.SocketInfo{
		Src_ip:   rand.Uint32(),
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

	go filterAndPublish(true, true, input, connections, servers)

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

	go filterAndPublish(true, true, input, connections, servers)

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

	go filterAndPublish(true, true, input, connections, servers)

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
