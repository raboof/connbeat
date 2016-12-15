package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/raboof/connbeat/processes"
	"github.com/raboof/connbeat/sockets"
	"github.com/raboof/connbeat/sockets/docker"
	"github.com/raboof/connbeat/sockets/proc_net_tcp"
	"github.com/raboof/connbeat/sockets/tcp_diag"
)

type ServerConnection struct {
	localIP   string
	localPort uint16
	process   *processes.UnixProcess
}

type Connection struct {
	localIP    string
	localPort  uint16
	remoteIp   string
	remotePort uint16
	process    *processes.UnixProcess
}

func getSocketInfoFromDocker(endpoint string, pollInterval time.Duration, socketInfo chan<- *sockets.SocketInfo) {
	for {
		// For now we poll periodically
		err := docker.PollCurrentConnections(endpoint, socketInfo)
		if err != nil {
			logp.Err("Polling connections: %s", err)
		}
		time.Sleep(pollInterval)
	}
}

func getSocketInfoFromProc(pollInterval time.Duration, socketInfo chan<- *sockets.SocketInfo) {
	for {
		// For now we poll periodically
		err := proc_net_tcp.PollCurrentConnections(socketInfo)
		if err != nil {
			logp.Err("Polling connections: %s", err)
		}
		time.Sleep(pollInterval)
	}
}

func getSocketInfoFromTcpDiag(pollInterval time.Duration, socketInfo chan<- *sockets.SocketInfo) {
	err := tcp_diag.GetSocketInfo(pollInterval, socketInfo)

	if err != nil {
		logp.Info("tcp_diag failed, falling back to /proc/net/tcp")
		getSocketInfoFromProc(pollInterval, socketInfo)
	}
}

func getSocketInfo(enableDocker, enableTcpDiag bool, pollInterval time.Duration, socketInfo chan<- *sockets.SocketInfo) {
	if enableDocker {
		getSocketInfoFromDocker("unix:///var/run/docker.sock", pollInterval, socketInfo)
	} else if enableTcpDiag {
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

func filterAndPublish(exposeProcessInfo, exposeCmdline, exposeEnviron bool, aggregation time.Duration, socketInfo <-chan *sockets.SocketInfo, connections chan<- Connection, servers chan ServerConnection) {
	listeningOn := make(map[uint16]time.Time)
	outgoingConnectionSeen := make(map[outgoingConnectionDedup]time.Time)
	ps := processes.New(exposeCmdline, exposeEnviron)

	for {
		now := time.Now()
		select {
		case s := <-socketInfo:
			if when, seen := listeningOn[s.SrcPort]; !seen || now.Sub(when) > aggregation {
				if s.DstPort == 0 {
					listeningOn[s.SrcPort] = now
					servers <- ServerConnection{
						localIP:   s.SrcIP.String(),
						localPort: s.SrcPort,
						process:   process(ps, exposeProcessInfo, s.Inode),
					}
				} else {
					dstIP := s.DstIP.String()
					dedupId := outgoingConnectionDedup{dstIP, s.DstPort}
					if when, seen := outgoingConnectionSeen[dedupId]; !seen || now.Sub(when) > aggregation {
						outgoingConnectionSeen[dedupId] = now
						connections <- Connection{
							localIP:    s.SrcIP.String(),
							localPort:  s.SrcPort,
							remoteIp:   dstIP,
							remotePort: s.DstPort,
							process:    process(ps, exposeProcessInfo, s.Inode),
						}
					}
				}
			}
		}
	}
}

func Listen(exposeProcessInfo, exposeCmdline, exposeEnviron, enableDocker, enableTcpDiag bool, pollInterval, aggregation time.Duration) (chan Connection, chan ServerConnection) {
	socketInfo := make(chan *sockets.SocketInfo, 20)

	go getSocketInfo(enableDocker, enableTcpDiag, pollInterval, socketInfo)

	connections := make(chan Connection, 20)
	servers := make(chan ServerConnection, 20)
	go filterAndPublish(exposeProcessInfo, exposeCmdline, exposeEnviron, aggregation, socketInfo, connections, servers)

	return connections, servers
}
