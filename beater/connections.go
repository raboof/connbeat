package beater

import (
	"time"
	"os"
	"fmt"

	"github.com/elastic/beats/packetbeat/procs"
	"github.com/elastic/beats/libbeat/logp"
)

type ServerConnection struct {
	localPort uint16
	process string
}

type Connection struct {
	localIp    string
	localPort  uint16
	remoteIp   string
	remotePort uint16
	process    string
}

func pollCurrentConnections(c chan Connection) {
	// TODO add support for IPv6
	// TODO add support for darwin
	file, err := os.Open("/proc/net/tcp")
	if err != nil {
		logp.Err("Open: %s", err)
		return
	}
	defer file.Close()
	socks, err := procs.Parse_Proc_Net_Tcp(file)
	for _, s := range socks {
		c <- Connection{
			localIp: formatIp(s.Src_ip),
			localPort: s.Src_port,
			remoteIp: formatIp(s.Dst_ip),
			remotePort: s.Dst_port,
			process: "nginx",
		}
	}
}

func formatIp(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24))
}

func bangOn(c chan Connection) {
	for true {
		// TODO use a scheduler instead of sleeping
		time.Sleep(2 * time.Second)
		pollCurrentConnections(c)
	}
}

func Listen() (chan Connection, chan ServerConnection) {
	connections := make(chan Connection, 20)
	servers := make(chan ServerConnection, 20)

	go bangOn(connections)

	return connections, servers
}
