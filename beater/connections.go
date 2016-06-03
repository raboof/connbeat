package beater

import (
	"fmt"
	"os"
	"time"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/packetbeat/procs"
	"github.com/raboof/connbeat/processes"
)

type ServerConnection struct {
	localIp   string
	localPort uint16
	process   string
}

type Connection struct {
	localIp    string
	localPort  uint16
	remoteIp   string
	remotePort uint16
	process    string
}

func getEnv(key, defaultValue string) string {
	env := os.Getenv(key)
	if env != "" {
		return env
	}
	return defaultValue
}

func pollCurrentConnections(socketInfo chan<- *procs.SocketInfo) {
	// TODO add support for IPv6
	// TODO add support for darwin
	// TODO prefer tcp_diag where available
	file, err := os.Open(getEnv("PROC_NET_TCP", "/proc/net/tcp"))
	if err != nil {
		logp.Err("Open: %s", err)
		return
	}
	defer file.Close()
	// TODO error handling
	socks, _ := procs.Parse_Proc_Net_Tcp(file)
	for _, s := range socks {
		socketInfo <- s
	}
}

func formatIp(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24))
}

func getSocketInfo(socketInfo chan<- *procs.SocketInfo) {
	for true {
		// For now we poll periodically, eventually we want to also be triggered on-demand
		time.Sleep(2 * time.Second)
		pollCurrentConnections(socketInfo)
	}
}

func filterAndPublish(socketInfo <-chan *procs.SocketInfo, connections chan<- Connection, servers chan ServerConnection) {
	listeningOn := make(map[uint16]bool)
	ps := processes.New()

	for {
		select {
		case s := <-socketInfo:
			if !listeningOn[s.Src_port] {
				if s.Dst_port == 0 {
					listeningOn[s.Src_port] = true
					servers <- ServerConnection{
						localIp:   formatIp(s.Src_ip),
						localPort: s.Src_port,
						process:   ps.FindProcessByInode(s.Inode),
					}
				} else {
					connections <- Connection{
						localIp:    formatIp(s.Src_ip),
						localPort:  s.Src_port,
						remoteIp:   formatIp(s.Dst_ip),
						remotePort: s.Dst_port,
						process:    ps.FindProcessByInode(s.Inode),
					}
				}
			}
		}
	}
}

func Listen() (chan Connection, chan ServerConnection) {
	socketInfo := make(chan *procs.SocketInfo, 20)

	go getSocketInfo(socketInfo)

	connections := make(chan Connection, 20)
	servers := make(chan ServerConnection, 20)
	go filterAndPublish(socketInfo, connections, servers)

	return connections, servers
}
