package beater

import (
	"strings"
	"time"

	"github.com/deckarep/golang-set"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/fsouza/go-dockerclient"

	"github.com/raboof/connbeat/processes"
	"github.com/raboof/connbeat/sockets"
)

type Connbeat struct {
	events     publisher.Client
	ConnConfig ConfigSettings

	done chan struct{}
}

type ContainerInfo struct {
	id          string
	localIPs    mapset.Set
	environment []string
	hostName    string
	ports       map[docker.Port][]docker.PortBinding
}

func New(b *beat.Beat, rawConfig *common.Config) (beat.Beater, error) {
	rawConnbeatConfig := rawConfig
	cb := &Connbeat{}
	err := cb.init(b)
	if err != nil {
		return nil, err
	}

	cb.ConnConfig.Connbeat = defaultConfig
	err = rawConnbeatConfig.Unpack(&cb.ConnConfig.Connbeat)
	if err != nil {
		logp.Err("Error reading configuration file: %v", err)
		return nil, err
	}

	logp.Debug("connbeat", "Expose process information: %v", cb.ConnConfig.Connbeat.ExposeProcessInfo)
	logp.Debug("connbeat", "Expose cmdline: %v", cb.ConnConfig.Connbeat.ExposeCmdline)
	logp.Debug("connbeat", "Expose environ: %v", cb.ConnConfig.Connbeat.ExposeEnviron)
	logp.Debug("connbeat", "Connection aggregation: %v", cb.ConnConfig.Connbeat.ConnectionAggregation)
	logp.Debug("connbeat", "Poll Interval %v", cb.ConnConfig.Connbeat.PollInterval)
	logp.Debug("connbeat", "Enable tcp_diag %v", cb.ConnConfig.Connbeat.TcpDiagEnabled)

	return cb, nil
}

func (cb *Connbeat) init(b *beat.Beat) error {
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

func bindingMap(bindings []docker.PortBinding) []common.MapStr {
	result := make([]common.MapStr, len(bindings))
	for _, binding := range bindings {
		result = append(result, common.MapStr{
			"HostIp":   binding.HostIP,
			"HostPort": binding.HostPort,
		})
	}
	return result
}

func portsMap(ports map[docker.Port][]docker.PortBinding) []common.MapStr {
	result := make([]common.MapStr, len(ports))
	for port, binding := range ports {
		result = append(result, common.MapStr{
			port.Port(): bindingMap(binding),
		})
	}
	return result
}

func toMap(containerInfo *ContainerInfo) common.MapStr {
	if containerInfo != nil {
		return common.MapStr{
			"id":        containerInfo.id,
			"local_ips": containerInfo.localIPs.ToSlice(),
			"env":       containerInfo.environment,
			"docker_host": common.MapStr{
				"name": containerInfo.hostName,
			},
			"ports": portsMap(containerInfo.ports),
		}
	}
	return nil
}

func (cb *Connbeat) exportServerConnection(s ServerConnection, localIPs mapset.Set, containerInfo *ContainerInfo) error {
	event := common.MapStr{
		"@timestamp":    common.Time(time.Now()),
		"type":          "connbeat",
		"local_port":    s.localPort,
		"local_process": processAsMap(s.process),
		"container":     toMap(containerInfo),
		"beat": common.MapStr{
			"local_ips": localIPs.ToSlice(),
		},
	}

	cb.events.PublishEvent(event)

	return nil
}

func (cb *Connbeat) exportConnection(c Connection, localIPs mapset.Set, containerInfo *ContainerInfo) error {
	event := common.MapStr{
		"@timestamp":    common.Time(time.Now()),
		"type":          "connbeat",
		"local_ip":      c.localIP,
		"local_port":    c.localPort,
		"remote_ip":     c.remoteIp,
		"remote_port":   c.remotePort,
		"local_process": processAsMap(c.process),
		"container":     toMap(containerInfo),
		"beat": common.MapStr{
			"local_ips": localIPs.ToSlice(),
		},
	}

	cb.events.PublishEvent(event)

	return nil
}

func update(infos map[string]ContainerInfo, socketContainerInfo *sockets.ContainerInfo, ip string) *ContainerInfo {
	if socketContainerInfo == nil {
		return nil
	}

	info, found := infos[socketContainerInfo.ID]
	if found {
		info.localIPs.Add(ip)
		return &info
	} else {
		localIPs := mapset.NewSet()
		localIPs.Add(ip)
		result := ContainerInfo{socketContainerInfo.ID, localIPs, socketContainerInfo.DockerEnvironment, socketContainerInfo.HostName, socketContainerInfo.Ports}
		infos[socketContainerInfo.ID] = result
		return &result
	}
}

func (cb *Connbeat) Pipe(connectionListener <-chan Connection, serverConnectionListener <-chan ServerConnection) error {
	var err error

	localIPs := mapset.NewSet()
	containerInfo := make(map[string]ContainerInfo)

	for {
		select {
		case <-cb.done:
			return nil
		case c := <-connectionListener:
			localIPs.Add(c.localIP)
			container := update(containerInfo, c.container, c.localIP)
			err = cb.exportConnection(c, localIPs, container)
			if err != nil {
				return err
			}
		case s := <-serverConnectionListener:
			if s.localIP != "0.0.0.0" &&
				s.localIP != "::" {
				localIPs.Add(s.localIP)
			}
			container := update(containerInfo, s.container, s.localIP)
			err = cb.exportServerConnection(s, localIPs, container)
			if err != nil {
				return err
			}
		}
	}
}

func (cb *Connbeat) Run(b *beat.Beat) error {
	connectionListener, serverConnectionListener, err := Listen(
		cb.ConnConfig.Connbeat.ExposeProcessInfo, cb.ConnConfig.Connbeat.ExposeCmdline, cb.ConnConfig.Connbeat.ExposeEnviron,
		cb.ConnConfig.Connbeat.DockerEnabled, cb.ConnConfig.Connbeat.TcpDiagEnabled,
		cb.ConnConfig.Connbeat.PollInterval, cb.ConnConfig.Connbeat.ConnectionAggregation,
		cb.ConnConfig.Connbeat.DockerEnvironment)

	if err != nil {
		return err
	}

	return cb.Pipe(connectionListener, serverConnectionListener)
}

func (cb *Connbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (cb *Connbeat) Stop() {
	close(cb.done)
}
