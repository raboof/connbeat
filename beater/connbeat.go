package beater

import (
	"time"

	"github.com/deckarep/golang-set"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"
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

func (cb *Connbeat) exportServerConnection(s ServerConnection, localIps mapset.Set) error {
	event := common.MapStr{
		"@timestamp":    common.Time(time.Now()),
		"type":          "connbeat",
		"local_port":    s.localPort,
		"local_process": s.process,
		"beat": common.MapStr{
			"local_ips": localIps.ToSlice(),
		},
	}

	cb.events.PublishEvent(event)

	return nil
}

func (cb *Connbeat) exportConnection(c Connection, localIps mapset.Set) error {
	event := common.MapStr{
		"@timestamp":    common.Time(time.Now()),
		"type":          "connbeat",
		"local_ip":      c.localIp,
		"local_port":    c.localPort,
		"remote_ip":     c.remoteIp,
		"remote_port":   c.remotePort,
		"local_process": c.process,
		"beat": common.MapStr{
			"local_ips": localIps.ToSlice(),
		},
	}

	cb.events.PublishEvent(event)

	return nil
}

func (cb *Connbeat) Pipe(connectionListener <-chan Connection, serverConnectionListener <-chan ServerConnection) error {
	var err error

	localIps := mapset.NewSet()

	for {
		select {
		case <-cb.done:
			return nil
		case c := <-connectionListener:
			localIps.Add(c.localIp)
			err = cb.exportConnection(c, localIps)
		case s := <-serverConnectionListener:
			if s.localIp != "0.0.0.0" {
				localIps.Add(s.localIp)
			}
			err = cb.exportServerConnection(s, localIps)
		}
	}

	return err
}

func (cb *Connbeat) Run(b *beat.Beat) error {
	connectionListener, serverConnectionListener := Listen()

	return cb.Pipe(connectionListener, serverConnectionListener)
}

func (cb *Connbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (cb *Connbeat) Stop() {
	close(cb.done)
}
