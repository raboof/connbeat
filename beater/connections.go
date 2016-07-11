package beater

import (
	"os"
	"time"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/packetbeat/procs"
	"github.com/raboof/connbeat/processes"
)

type ServerConnection struct {
	localIp   string
	localPort uint16
	process   *processes.UnixProcess
}

type Connection struct {
	localIp    string
	localPort  uint16
	remoteIp   string
	remotePort uint16
	process    *processes.UnixProcess
}

func getEnv(key, defaultValue string) string {
	env := os.Getenv(key)
	if env != "" {
		return env
	}
	return defaultValue
}

func pollCurrentConnections(socketInfo chan<- *procs.SocketInfo) error {
	// TODO add support for darwin
	// TODO prefer tcp_diag where available
	err := pollCurrentConnectionsFrom(getEnv("PROC_NET_TCP", "/proc/net/tcp"), false, socketInfo)
	if err != nil {
		return err
	}
	return pollCurrentConnectionsFrom(getEnv("PROC_NET_TCP6", "/proc/net/tcp6"), true, socketInfo)
}

func pollCurrentConnectionsFrom(filename string, ipv6 bool, socketInfo chan<- *procs.SocketInfo) error {
	file, err := os.Open(filename)
	if err != nil {
		logp.Err("Open: %s", err)
		return err
	}
	defer file.Close()
	socks, err := procs.Parse_Proc_Net_Tcp(file, ipv6)
	if err != nil {
		return err
	}
	for _, s := range socks {
		if s.Inode != 0 {
			socketInfo <- s
		}
	}
	return nil
}

func getSocketInfo(socketInfo chan<- *procs.SocketInfo) {
	for true {
		// For now we poll periodically, eventually we want to also be triggered on-demand
		time.Sleep(2 * time.Second)
		err := pollCurrentConnections(socketInfo)
		if err != nil {
			logp.Err("Polling connections: %s", err)
		}
	}
}

type outgoingConnectionDedup struct {
	remoteIp   string
	remotePort uint16
}

func filterAndPublish(exposeCmdline, exposeEnviron bool, aggregation time.Duration, socketInfo <-chan *procs.SocketInfo, connections chan<- Connection, servers chan ServerConnection) {
	listeningOn := make(map[uint16]time.Time)
	outgoingConnectionSeen := make(map[outgoingConnectionDedup]time.Time)
	ps := processes.New(exposeCmdline, exposeEnviron)

	for {
		now := time.Now()
		select {
		case s := <-socketInfo:
			if when, seen := listeningOn[s.Src_port]; !seen || now.Sub(when) > aggregation {
				if s.Dst_port == 0 {
					listeningOn[s.Src_port] = now
					servers <- ServerConnection{
						localIp:   s.Src_ip.String(),
						localPort: s.Src_port,
						process:   ps.FindProcessByInode(s.Inode),
					}
				} else {
					dstIp := s.Dst_ip.String()
					dedupId := outgoingConnectionDedup{dstIp, s.Dst_port}
					if when, seen := outgoingConnectionSeen[dedupId]; !seen || now.Sub(when) > aggregation {
						outgoingConnectionSeen[dedupId] = now
						connections <- Connection{
							localIp:    s.Src_ip.String(),
							localPort:  s.Src_port,
							remoteIp:   dstIp,
							remotePort: s.Dst_port,
							process:    ps.FindProcessByInode(s.Inode),
						}
					}
				}
			}
		}
	}
}

func Listen(exposeCmdline, exposeEnviron bool, aggregation time.Duration) (chan Connection, chan ServerConnection) {
	socketInfo := make(chan *procs.SocketInfo, 20)

	go getSocketInfo(socketInfo)

	connections := make(chan Connection, 20)
	servers := make(chan ServerConnection, 20)
	go filterAndPublish(exposeCmdline, exposeEnviron, aggregation, socketInfo, connections, servers)

	return connections, servers
}
