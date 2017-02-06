package connections

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

type LocalConnection struct {
	LocalIP   string
	LocalPort uint16

	Process   *processes.UnixProcess
	Container *sockets.ContainerInfo
}

type ServerConnection LocalConnection

type FullConnection struct {
	LocalConnection

	RemoteIP   string
	RemotePort uint16
}

type Connections struct {
	exposeProcessInfo      bool
	listeningOn            map[incomingConnectionDedup]time.Time
	outgoingConnectionSeen map[outgoingConnectionDedup]time.Time
	ps                     *processes.Processes
}

func New(exposeCmdline, exposeEnviron bool) *Connections {
	return &Connections{
		listeningOn:            make(map[incomingConnectionDedup]time.Time),
		outgoingConnectionSeen: make(map[outgoingConnectionDedup]time.Time),
		ps: processes.New(exposeCmdline, exposeEnviron),
	}
}

func getSocketInfoFromDocker(poller *docker.Poller, pollInterval time.Duration, socketInfo chan<- *sockets.SocketInfo) {
	for {
		logp.Info("Polling docker")
		// For now we poll periodically
		err := poller.PollCurrentConnections(socketInfo)
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

func getSocketInfo(enableTcpDiag bool, pollInterval time.Duration, socketInfo chan<- *sockets.SocketInfo) {
	if enableTcpDiag {
		getSocketInfoFromTcpDiag(pollInterval, socketInfo)
	} else {
		getSocketInfoFromProc(pollInterval, socketInfo)
	}
}

type incomingConnectionDedup struct {
	localIp   string
	localPort uint16
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

func (c *Connections) filterAndPublish(exposeProcessInfo bool, aggregation time.Duration, socketInfo <-chan *sockets.SocketInfo, connections chan<- FullConnection, servers chan ServerConnection) {
	for {
		now := time.Now()
		select {
		case s := <-socketInfo:
			localIP := s.SrcIP.String()
			localDedupId := incomingConnectionDedup{localIP, s.SrcPort}
			logp.Info("connections", "considering connection %v", s)
			if when, seen := c.listeningOn[localDedupId]; !seen || now.Sub(when) > aggregation {
				c.listeningOn[localDedupId] = now
				logp.Info("connections", "really considering connection %v", s)
				if s.DstPort == 0 {
					servers <- ServerConnection{
						LocalIP:   localIP,
						LocalPort: s.SrcPort,
						Process:   process(c.ps, c.exposeProcessInfo && s.Container == nil, s.Inode),
						Container: s.Container,
					}
				} else {
					dstIP := s.DstIP.String()
					dedupId := outgoingConnectionDedup{dstIP, s.DstPort}
					if when, seen := c.outgoingConnectionSeen[dedupId]; !seen || now.Sub(when) > aggregation {
						c.outgoingConnectionSeen[dedupId] = now
						logp.Info("connections", "adding connection %v", s)
						connections <- FullConnection{
							LocalConnection{
								LocalIP:   s.SrcIP.String(),
								LocalPort: s.SrcPort,
								Process:   process(c.ps, exposeProcessInfo && s.Container == nil, s.Inode),
								Container: s.Container,
							},
							dstIP,
							s.DstPort,
						}
					}
				}
			}
		}
	}
}

func Listen(exposeProcessInfo, exposeCmdline, exposeEnviron,
	enableLocalConnections, enableDocker, enableTcpDiag bool,
	pollInterval, aggregation time.Duration,
	dockerEnvironment []string) (chan FullConnection, chan ServerConnection, error) {
	socketInfo := make(chan *sockets.SocketInfo, 20)

	if enableDocker {
		poller, err := docker.New(dockerEnvironment)
		if err != nil {
			return nil, nil, err
		}
		go getSocketInfoFromDocker(poller, pollInterval, socketInfo)
	}
	if enableLocalConnections {
		go getSocketInfo(enableTcpDiag, pollInterval, socketInfo)
	}

	connections := make(chan FullConnection, 20)
	servers := make(chan ServerConnection, 20)
	go New(exposeProcessInfo, exposeCmdline).filterAndPublish(exposeEnviron, aggregation, socketInfo, connections, servers)

	return connections, servers, nil
}
