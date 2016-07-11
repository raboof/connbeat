package beater

import (
	"strings"
	"time"

	"github.com/deckarep/golang-set"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/raboof/connbeat/processes"
)

type Connbeat struct {
	events     publisher.Client
	ConnConfig ConfigSettings

	done chan struct{}
}

func New() *Connbeat {
	return &Connbeat{}
}

func (cb *Connbeat) Config(b *beat.Beat) error {
	rawConnbeatConfig, err := b.RawConfig.Child("connbeat", -1)
	if err != nil {
		logp.Err("Error reading configuration file: %v", err)
		return err
	}

	cb.ConnConfig.Connbeat = defaultConfig
	err = rawConnbeatConfig.Unpack(&cb.ConnConfig.Connbeat)
	if err != nil {
		logp.Err("Error reading configuration file: %v", err)
		return err
	}

	logp.Debug("connbeat", "Expose cmdline: %v", cb.ConnConfig.Connbeat.ExposeCmdline)
	logp.Debug("connbeat", "Expose environ: %v", cb.ConnConfig.Connbeat.ExposeEnviron)
	logp.Debug("connbeat", "Connection aggregation: %v", cb.ConnConfig.Connbeat.ConnectionAggregation)

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

func (cb *Connbeat) exportServerConnection(s ServerConnection, localIps mapset.Set) error {
	event := common.MapStr{
		"@timestamp":    common.Time(time.Now()),
		"type":          "connbeat",
		"local_port":    s.localPort,
		"local_process": processAsMap(s.process),
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
		"local_process": processAsMap(c.process),
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
			if err != nil {
				return err
			}
		case s := <-serverConnectionListener:
			if s.localIp != "0.0.0.0" {
				localIps.Add(s.localIp)
			}
			err = cb.exportServerConnection(s, localIps)
			if err != nil {
				return err
			}
		}
	}
}

func (cb *Connbeat) Run(b *beat.Beat) error {
	connectionListener, serverConnectionListener := Listen(cb.ConnConfig.Connbeat.ExposeCmdline, cb.ConnConfig.Connbeat.ExposeEnviron, cb.ConnConfig.Connbeat.ConnectionAggregation)

	return cb.Pipe(connectionListener, serverConnectionListener)
}

func (cb *Connbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (cb *Connbeat) Stop() {
	close(cb.done)
}
