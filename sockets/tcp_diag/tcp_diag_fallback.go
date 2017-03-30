// +build !linux

package tcp_diag

import (
	"errors"
	"time"

	"github.com/raboof/connbeat/sockets"
)

func GetSocketInfo(pollInterval time.Duration, socketInfo chan<- *sockets.SocketInfo) error {
	return errors.New("tcp_diag is only supported on linux")
}
