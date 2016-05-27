package beater

import (
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/raboof/connbeat/processes"
)

type Connbeat struct {
	events publisher.Client

	done chan struct{}
}

func New() *Connbeat {
	return &Connbeat{}
}

func (cb *Connbeat) Config(b *beat.Beat) error {
	return nil
}

func (cb *Connbeat) Setup(b *beat.Beat) error {
	cb.events = b.Publisher.Connect()
	cb.done = make(chan struct{})
	return nil
}

func processAsMap(process *processes.UnixProcess) common.MapStr {
	binary := strings.Trim(process.Binary, "\u0000 ")
	cmdline := strings.Trim(strings.Replace(process.Cmdline, "\u0000", " ", -1), "\u0000 ")
	environ := strings.Split(strings.Trim(process.Environ, "\u0000 "), "\u0000")
	return common.MapStr{
		"binary":  binary,
		"cmdline": cmdline,
		"environ": environ,
	}
}

func (cb *Connbeat) exportServerConnection(s ServerConnection) error {
	event := common.MapStr{
		"@timestamp":    common.Time(time.Now()),
		"type":          "connbeat",
		"local_port":    s.localPort,
		"local_process": processAsMap(s.process),
	}

	cb.events.PublishEvent(event)

	return nil
}

func (cb *Connbeat) exportConnection(c Connection) error {
	event := common.MapStr{
		"@timestamp":    common.Time(time.Now()),
		"type":          "connbeat",
		"local_ip":      c.localIp,
		"local_port":    c.localPort,
		"remote_ip":     c.remoteIp,
		"remote_port":   c.remotePort,
		"local_process": processAsMap(c.process),
	}

	cb.events.PublishEvent(event)

	return nil
}

func (cb *Connbeat) Run(b *beat.Beat) error {
	var err error

	connectionListener, serverConnectionListener := Listen()

	for {
		select {
		case <-cb.done:
			return nil
		case c := <-connectionListener:
			err = cb.exportConnection(c)
		case s := <-serverConnectionListener:
			err = cb.exportServerConnection(s)
		}
	}

	return err
}

func (cb *Connbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (cb *Connbeat) Stop() {
	close(cb.done)
}
