package connections

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
	connections, servers := make(chan FullConnection, 0), make(chan ServerConnection, 0)

	go New(true, true).filterAndPublish(true, 5*time.Second, input, connections, servers)

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
	connections, servers := make(chan FullConnection, 0), make(chan ServerConnection, 0)

	go New(true, true).filterAndPublish(true, 5*time.Second, input, connections, servers)

	remoteIP := randIP()

	input <- incomingConnectionTo(80, remoteIP)
	_, ok := <-connections
	assert.True(t, ok, "a server should be reported")
	input <- incomingConnectionTo(80, randIP())
	_, ok = <-connections
	assert.True(t, ok, "a server with a different local IP should be reported")
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
	connections, servers := make(chan FullConnection, 0), make(chan ServerConnection, 0)

	go New(true, true).filterAndPublish(true, 5*time.Second, input, connections, servers)

	localIP := randIP()

	input <- listeningConnectionOn(80, localIP)
	_, ok := <-servers
	assert.True(t, ok, "a server should be reported")
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

func TestDedupIdenticalClientConnections(t *testing.T) {
	input := make(chan *sockets.SocketInfo, 0)
	connections, servers := make(chan FullConnection, 0), make(chan ServerConnection, 0)

	go New(true, true).filterAndPublish(true, 5*time.Second, input, connections, servers)

	remoteIP := randIP()
	conn := outgoingConnection(remoteIP, 80)
	input <- conn
	_, ok := <-connections
	assert.True(t, ok, "a client connection should be reported")
	input <- conn

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

func TestReportDistinctClientConnectionsToTheSameServer(t *testing.T) {
	input := make(chan *sockets.SocketInfo, 0)
	connections, servers := make(chan FullConnection, 0), make(chan ServerConnection, 0)

	go New(true, true).filterAndPublish(true, 5*time.Second, input, connections, servers)

	remoteIP := randIP()
	input <- outgoingConnection(remoteIP, 80)
	_, ok := <-connections
	assert.True(t, ok, "a client connection should be reported")
	input <- outgoingConnection(remoteIP, 80)
	_, ok = <-connections
	assert.True(t, ok, "another client connection to the same server should be reported")
}

func TestRepublishOldClientConnections(t *testing.T) {
	input := make(chan *sockets.SocketInfo, 0)
	connections, servers := make(chan FullConnection, 0), make(chan ServerConnection, 0)

	go New(false, true).filterAndPublish(false, 0*time.Second, input, connections, servers)

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
	connections, servers := make(chan FullConnection, 100), make(chan ServerConnection, 100)

	go New(false, true).filterAndPublish(false, 10*time.Second, input, connections, servers)

	insert174data(t, input)

	expectConnectionOnPort(t, connections, 35074)
}

func expectConnectionOnPort(t *testing.T, connections <-chan FullConnection, port uint16) {
	done := make(chan string, 1)

	go func() {
		time.Sleep(2 * time.Second)
		done <- "timeout"
	}()

	for {
		select {
		case connection := <-connections:
			if connection.LocalPort == port {
				t.Log("Found connection", connection)
				done <- "found"
			} else {
				t.Log("Ignored connection", connection)
			}
		case reason := <-done:
			if reason == "found" {
				// OK!
			} else if reason == "timeout" {
				t.Fatal("Did not find connection with local port", port)
			} else {
				t.Fatal("Unexpected reason", reason)
			}
			return
		}
	}
}

func insert174data(t *testing.T, socketInfo chan<- *sockets.SocketInfo) {
	file, err := os.Open("../tests/files/proc_net_tcp6_174.txt")
	if err != nil {
		t.Fatalf("Opening ../../tests/files/proc_net_tcp6_174.txt: %s", err)
	}
	sockets, err := proc_net_tcp.ParseProcNetTCP(file, true, nil)
	if err != nil {
		t.Fatalf("Parsing ../../tests/files/proc_net_tcp6_174.txt: %s", err)
	}
	for _, socket := range sockets {
		socketInfo <- socket
	}
}
