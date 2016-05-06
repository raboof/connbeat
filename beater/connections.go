package beater

import (
	"time"
	// "github.com/weaveworks/scope/probe/endpoint"
)

type Connection struct {
	localIp    string
	localPort  int
	remoteIp   string
	remotePort int
	process    string
}

func bangOn(c chan Connection) {
	for true {
		time.Sleep(2 * time.Second)
		c <- Connection{
			localIp:    "132.32.1.3",
			localPort:  32,
			remoteIp:   "43.34.1.3",
			remotePort: 1243,
			process:    "ftpd",
		}
	}
}

func Listen() chan Connection {
	result := make(chan Connection, 20)
	// TODO actually start monitoring
	go bangOn(result)
	return result
}
