package proc_net_tcp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/raboof/connbeat/sockets"
)

func getEnv(key, defaultValue string) string {
	env := os.Getenv(key)
	if env != "" {
		return env
	}
	return defaultValue
}

func PollCurrentConnections(socketInfo chan<- *sockets.SocketInfo) error {
	// TODO add support for darwin
	// TODO prefer tcp_diag where available
	err := pollCurrentConnectionsFrom(getEnv("PROC_NET_TCP", "/proc/net/tcp"), false, socketInfo)
	if err != nil {
		return err
	}
	return pollCurrentConnectionsFrom(getEnv("PROC_NET_TCP6", "/proc/net/tcp6"), true, socketInfo)
}

func hexToIPPort(str []byte, ipv6 bool) (net.IP, uint16, error) {
	words := bytes.Split(str, []byte(":"))
	if len(words) < 2 {
		return nil, 0, errors.New("Didn't find ':' as a separator")
	}

	ip, err := hexToIP(string(words[0]), ipv6)
	if err != nil {
		return nil, 0, err
	}

	port, err := strconv.ParseInt(string(words[1]), 16, 32)
	if err != nil {
		return nil, 0, err
	}

	return ip, uint16(port), nil
}

func hexToIpv4(word string) (net.IP, error) {
	ip, err := strconv.ParseInt(word, 16, 64)
	if err != nil {
		return nil, err
	}
	return net.IPv4(byte(ip), byte(ip>>8), byte(ip>>16), byte(ip>>24)), nil
}

func hexToIpv6(word string) (net.IP, error) {
	p := make(net.IP, net.IPv6len)
	for i := 0; i < 4; i++ {
		part, err := strconv.ParseInt(word[i*8:(i+1)*8], 16, 32)
		if err != nil {
			return nil, err
		}
		p[i*4] = byte(part)
		p[i*4+1] = byte(part >> 8)
		p[i*4+2] = byte(part >> 16)
		p[i*4+3] = byte(part >> 24)
	}
	return p, nil
}

func hexToIP(word string, ipv6 bool) (net.IP, error) {
	if ipv6 {
		return hexToIpv6(word)
	}
	return hexToIpv4(word)
}

// Parses the /proc/net/tcp file
func ParseProcNetTCP(input io.Reader, ipv6 bool, containerId string) ([]*sockets.SocketInfo, error) {
	buf := bufio.NewReader(input)

	result := []*sockets.SocketInfo{}
	var err error
	var line []byte
	for err != io.EOF {
		line, err = buf.ReadBytes('\n')
		if err != nil && err != io.EOF {
			logp.Err("Error reading /proc/net/tcp: %s", err)
			return nil, err
		}
		words := bytes.Fields(line)
		if len(words) < 10 || bytes.Equal(words[0], []byte("sl")) {
			logp.Debug("procs", "Less then 10 words (%d) or starting with 'sl': %s", len(words), words)
			continue
		}

		var sock sockets.SocketInfo
		sock.ContainerId = containerId
		var err error

		sock.SrcIP, sock.SrcPort, err = hexToIPPort(words[1], ipv6)
		if err != nil {
			logp.Debug("procs", "Error parsing IP and port: %s", err)
			continue
		}

		sock.DstIP, sock.DstPort, err = hexToIPPort(words[2], ipv6)
		if err != nil {
			logp.Debug("procs", "Error parsing IP and port: %s", err)
			continue
		}

		uid, _ := strconv.Atoi(string(words[7]))
		sock.UID = uint32(uid)
		inode, _ := strconv.Atoi(string(words[9]))
		sock.Inode = uint64(inode)

		result = append(result, &sock)
	}
	return result, nil
}

func pollCurrentConnectionsFrom(filename string, ipv6 bool, socketInfo chan<- *sockets.SocketInfo) error {
	file, err := os.Open(filename)
	if err != nil {
		logp.Err("Open: %s", err)
		return err
	}
	defer file.Close()
	socks, err := ParseProcNetTCP(file, ipv6, "")
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
