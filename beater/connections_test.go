package beater

import (
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/raboof/connbeat/sockets"
	"github.com/raboof/connbeat/sockets/proc_net_tcp"
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

func listeningConnection(port uint16) *sockets.SocketInfo {
	return listeningConnectionOn(port, randIP())
}

func listeningConnectionOn(localPort uint16, localIP net.IP) *sockets.SocketInfo {
	return &sockets.SocketInfo{
		SrcIP:   localIP,
		DstIP:   randIP(),
		SrcPort: localPort,
		DstPort: 0,
		UID:     uint32(rand.Int()),
		Inode:   uint64(rand.Int()),
	}
}

func incomingConnection(localPort uint16) *sockets.SocketInfo {
	return incomingConnectionTo(localPort, randIP())
}

func incomingConnectionTo(localPort uint16, localIP net.IP) *sockets.SocketInfo {
	return &sockets.SocketInfo{
		SrcIP:   localIP,
		DstIP:   randIP(),
		SrcPort: localPort,
		DstPort: uint16(rand.Int()),
		UID:     uint32(rand.Int()),
		Inode:   uint64(rand.Int63()),
	}
}

func outgoingConnection(remoteIP net.IP, remotePort uint16) *sockets.SocketInfo {
	return &sockets.SocketInfo{
		SrcIP:   randIP(),
		DstIP:   remoteIP,
		SrcPort: uint16(rand.Int()),
		DstPort: remotePort,
		UID:     uint32(rand.Int()),
		Inode:   uint64(rand.Int63()),
	}
}

func TestDeduplicateListeningSockets(t *testing.T) {
	input := make(chan *sockets.SocketInfo, 0)
	connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

	go filterAndPublish(true, true, true, 5*time.Second, input, connections, servers)

	ip := randIP()

	input <- listeningConnectionOn(80, ip)
	_, ok := <-servers
	assert.Equal(t, ok, true, "a server should be reported")
	input <- listeningConnectionOn(80, ip)

	time.Sleep(100 * time.Millisecond)

	select {
	case <-connections:
		t.Fatal("Saw a spurious connection")
	case <-servers:
		t.Fatal("Saw a spurious server connection")
	default:
		// Nothing to read: OK!
	}
}

func TestFilterIncomingConnectionsPerIP(t *testing.T) {
	input := make(chan *sockets.SocketInfo, 0)
	connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

	go filterAndPublish(true, true, true, 5*time.Second, input, connections, servers)

	remoteIP := randIP()

	input <- incomingConnectionTo(80, remoteIP)
	_, ok := <-connections
	assert.Equal(t, ok, true, "a server should be reported")
	input <- incomingConnectionTo(80, randIP())
	_, ok = <-connections
	assert.Equal(t, ok, true, "a server with a different local IP should be reported")
	input <- incomingConnectionTo(80, remoteIP)

	time.Sleep(100 * time.Millisecond)

	select {
	case <-connections:
		t.Fatal("a server for which we already saw the local IP should not be reported")
	case <-servers:
		t.Fatal("no server connections should be reported")
	default:
		// Nothing to read: OK!
	}
}

func TestFilterConnectionsAssociatedWithListeningSockets(t *testing.T) {
	input := make(chan *sockets.SocketInfo, 0)
	connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

	go filterAndPublish(true, true, true, 5*time.Second, input, connections, servers)

	localIP := randIP()

	input <- listeningConnectionOn(80, localIP)
	_, ok := <-servers
	assert.Equal(t, ok, true, "a server should be reported")
	input <- incomingConnectionTo(80, localIP)

	time.Sleep(100 * time.Millisecond)

	select {
	case <-connections:
		t.Fatal("The incoming connection on the known local IP should not be reported")
	case <-servers:
		t.Fatal("No server notification expected")
	default:
		// Nothing to read: OK!
	}
}

func TestDedupClientConnections(t *testing.T) {
	input := make(chan *sockets.SocketInfo, 0)
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
	input := make(chan *sockets.SocketInfo, 0)
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

func Test174(t *testing.T) {
	input := make(chan *sockets.SocketInfo, 0)
	connections, servers := make(chan Connection, 100), make(chan ServerConnection, 100)

	go filterAndPublish(false, false, true, 0*time.Second, input, connections, servers)

	insert174data(t, input)

	expectConnectionOnPort(t, connections, 35074)
}

func expectConnectionOnPort(t *testing.T, connections <-chan Connection, port uint16) {
	found := false
	for !found {
		select {
		case connection := <-connections:
			if connection.localPort == port {
				t.Log("Found connection %v", connection)
				found = true
			} else {
				t.Log("Ignored connection %v", connection)
			}
		default:
			if !found {
				t.Fatal("Did not find connection with local port %d", port)
			}
		}
	}

}

func insert174data(t *testing.T, socketInfo chan<- *sockets.SocketInfo) {
	file, err := os.Open("../tests/files/proc_net_tcp6_174.txt")
	if err != nil {
		t.Fatalf("Opening ../../tests/files/proc_net_tcp.txt: %s", err)
	}
	sockets, err := proc_net_tcp.ParseProcNetTCP(file, true, nil)
	if err != nil {
		t.Fatalf("Opening ../../tests/files/proc_net_tcp6_174.txt: %s", err)
	}
	for _, socket := range sockets {
		socketInfo <- socket
	}
}
