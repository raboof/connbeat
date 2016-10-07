package beater

import (
	"fmt"
	"os"
	"time"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/packetbeat/procs"
	"github.com/raboof/connbeat/processes"
	"github.com/raboof/connbeat/tcp_diag"
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

func getSocketInfoFromProc(pollInterval time.Duration, socketInfo chan<- *procs.SocketInfo) {
	for {
		// For now we poll periodically
		err := pollCurrentConnections(socketInfo)
		if err != nil {
			logp.Err("Polling connections: %s", err)
		}
		time.Sleep(pollInterval)
	}
}

func getSocketInfoFromTcpDiag(pollInterval time.Duration, socketInfo chan<- *procs.SocketInfo) {
	err := tcp_diag.GetSocketInfo(pollInterval, socketInfo)

	if err != nil {
		logp.Info("tcp_diag failed, falling back to /proc/net/tcp")
		getSocketInfoFromProc(pollInterval, socketInfo)
	}
}

func getSocketInfo(enableTcpDiag bool, pollInterval time.Duration, socketInfo chan<- *procs.SocketInfo) {
	if enableTcpDiag {
		getSocketInfoFromTcpDiag(pollInterval, socketInfo)
	} else {
		getSocketInfoFromProc(pollInterval, socketInfo)
	}
}

type outgoingConnectionDedup struct {
	remoteIp   string
	remotePort uint16
}

func process(ps *processes.Processes, exposeProcessInfo bool, inode uint64) *processes.UnixProcess {
	if exposeProcessInfo {
		proc := ps.FindProcessByInode(inode)
		if proc != nil {
			return proc
		}
		return &processes.UnixProcess{
			Binary: fmt.Sprintf("Unknown process with inode %d", inode),
		}
	} else {
		return &processes.UnixProcess{
			Binary: fmt.Sprintf("Process with inode %d", inode),
		}
	}
}

func filterAndPublish(exposeProcessInfo, exposeCmdline, exposeEnviron bool, aggregation time.Duration, socketInfo <-chan *procs.SocketInfo, connections chan<- Connection, servers chan ServerConnection) {
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
						process:   process(ps, exposeProcessInfo, s.Inode),
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
							process:    process(ps, exposeProcessInfo, s.Inode),
						}
					}
				}
			}
		}
	}
}

func Listen(exposeProcessInfo, exposeCmdline, exposeEnviron, enableTcpDiag bool, pollInterval, aggregation time.Duration) (chan Connection, chan ServerConnection) {
	socketInfo := make(chan *procs.SocketInfo, 20)

	go getSocketInfo(enableTcpDiag, pollInterval, socketInfo)

	connections := make(chan Connection, 20)
	servers := make(chan ServerConnection, 20)
	go filterAndPublish(exposeProcessInfo, exposeCmdline, exposeEnviron, aggregation, socketInfo, connections, servers)

	return connections, servers
}
