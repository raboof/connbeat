package beater

import (
  "time"
  "testing"
  "github.com/stvp/assert"
  "math/rand"

  "github.com/elastic/beats/packetbeat/procs"
)

func listeningConnection(port uint16) *procs.SocketInfo {
   return &procs.SocketInfo { rand.Uint32(), rand.Uint32(), port, 0, uint16(rand.Int()), rand.Int63() }
}

func incomingConnection(localPort uint16) *procs.SocketInfo {
   return &procs.SocketInfo { rand.Uint32(), rand.Uint32(), localPort, uint16(rand.Uint32()), uint16(rand.Int()), rand.Int63() }
}

func TestDeduplicateListeningSockets(t *testing.T) {
  input := make(chan *procs.SocketInfo, 0)
  connections, servers := make(chan Connection, 0), make(chan ServerConnection, 0)

  go filterAndPublish(input, connections, servers)

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

  go filterAndPublish(input, connections, servers)

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
