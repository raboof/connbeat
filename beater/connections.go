package beater

import (
	"fmt"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"net"
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

func parse_ip_port(str string) (net.IP, uint16, error) {
	words := strings.Split(str, ":")
	if len(words) < 2 {
		return nil, 0, errors.New("Didn't find ':' as a separator")
	}

	ip := net.ParseIP(words[0])

	port, err := strconv.ParseInt(words[1], 10, 32)

	if err != nil {
		return nil, 0, err
	}

	return ip, uint16(port), nil
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

func getSocketInfoFromNetstat(pollInterval time.Duration, socketInfo chan<- *procs.SocketInfo) {
	//attr := os.ProcAttr{}
	//attr.Sys.HideWindow = true
	for {
		// For now we poll periodically
		cmd := exec.Command("netstat", "-ano", "-p", "TCP")
    stdout, err := cmd.StdoutPipe()
		if err != nil {
			logp.Err("opening stdout: %s", err)
		}
		err = cmd.Start()
		if err != nil {
			logp.Err("starting netstat: %s", err)
		}
		buf := new(bytes.Buffer)
		// TODO eventually we'd want to stream this, for now just get the whole thing:
		buf.ReadFrom(stdout)
		lines := strings.Split(buf.String(), "\n")[4:]
		for _, l := range lines {
		  words := strings.Fields(l)
			if len(words) == 5 {
				var sock procs.SocketInfo
				var err_ error
				sock.Src_ip, sock.Src_port, err_ = parse_ip_port(words[1])
				if err_ != nil {
					logp.Debug("connections", "Error parsing src IP and port: %s", err_)
					continue
				}
				sock.Dst_ip, sock.Dst_port, err_ = parse_ip_port(words[2])
				if err_ != nil {
					logp.Debug("connections", "Error parsing dst IP and port: %s", err_)
					continue
				}
				pid, err_ := strconv.Atoi(words[4])
				if err != nil {
					logp.Debug("connections", "Error parsing pid: %s", err_)
					continue
				}
				// TODO For now we abuse the Uid+Inode fields to store the Pid...
				sock.Inode = int64(pid)
				sock.Uid = 65535
				socketInfo <- &sock
			}
		}
		time.Sleep(pollInterval)
	}
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
	getSocketInfoFromNetstat(pollInterval, socketInfo)
	// if enableTcpDiag {
	// 	getSocketInfoFromTcpDiag(pollInterval, socketInfo)
	// } else {
	// 	getSocketInfoFromProc(pollInterval, socketInfo)
	// }
}

type outgoingConnectionDedup struct {
	remoteIp   string
	remotePort uint16
}

func processByPid(ps *processes.Processes, pid int64) *processes.UnixProcess {
	proc := ps.FindProcessByPid(pid)
	if proc != nil {
		return proc
	}
	return &processes.UnixProcess{
		Binary: fmt.Sprintf("Unknown process with pid %d", pid),
	}
}

func processByInode(ps *processes.Processes, inode int64) *processes.UnixProcess {
	proc := ps.FindProcessByInode(inode)
	if proc != nil {
		return proc
	}
	return &processes.UnixProcess{
		Binary: fmt.Sprintf("Unknown process with inode %d", inode),
	}
}

func process(ps *processes.Processes, exposeProcessInfo bool, uid uint16, inode int64) *processes.UnixProcess {
	if exposeProcessInfo {
		if uid == 65535 {
			return processByPid(ps, inode)
		} else {
			return processByInode(ps, inode)
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
						process:   process(ps, exposeProcessInfo, s.Uid, s.Inode),
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
							process:    process(ps, exposeProcessInfo, s.Uid, s.Inode),
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
