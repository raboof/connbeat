package beater

import (
	"net"
	"strings"
	"time"

	"github.com/deckarep/golang-set"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/fsouza/go-dockerclient"

	"github.com/raboof/connbeat/connections"
	"github.com/raboof/connbeat/processes"
	"github.com/raboof/connbeat/sockets"
)

type Connbeat struct {
	events     publisher.Client
	ConnConfig ConfigSettings

	done chan struct{}
}

type ContainerInfo struct {
	id                 string
	name               string
	image              string
	localIPs           mapset.Set
	environment        []string
	labels             map[string]string
	ports              map[docker.Port][]docker.PortBinding
	dockerHostHostname string
	dockerHostIP       net.IP
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
	logp.Debug("connbeat", "Enable docker", cb.ConnConfig.Connbeat.DockerEnabled)
	logp.Debug("connbeat", "Enable local connections", cb.ConnConfig.Connbeat.LocalConnectionsEnabled)

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
	proc := common.MapStr{
		"binary":  binary,
		"cmdline": cmdline,
		"environ": environ,
	}
	pid := process.Pid()
	if pid != 0 {
		proc["pid"] = pid
	}

	return proc
}

func toIPs(ip net.IP) []net.IP {
	if ip == nil {
		return []net.IP{}
	} else {
		return []net.IP{ip}
	}
}

func bindingMap(bindings []docker.PortBinding) []common.MapStr {
	result := make([]common.MapStr, len(bindings))
	for idx, binding := range bindings {
		result[idx] = common.MapStr{
			"HostIp":   binding.HostIP,
			"HostPort": binding.HostPort,
		}
	}
	return result
}

func portsMap(ports map[docker.Port][]docker.PortBinding) common.MapStr {
	result := common.MapStr{}
	for port, binding := range ports {
		result[port.Port()] = bindingMap(binding)
	}
	return result
}

func labelsMap(labels map[string]string) common.MapStr {
	result := common.MapStr{}
	for key, value := range labels {
		result[key] = value
	}
	return result
}

func toMap(containerInfo *ContainerInfo) common.MapStr {
	if containerInfo != nil {
		return common.MapStr{
			"id":        containerInfo.id,
			"name":      containerInfo.name,
			"image":     containerInfo.image,
			"local_ips": containerInfo.localIPs.ToSlice(),
			"env":       containerInfo.environment,
			"labels":    labelsMap(containerInfo.labels),
			"docker_host": common.MapStr{
				"hostname": containerInfo.dockerHostHostname,
				"ips":      toIPs(containerInfo.dockerHostIP),
			},
			"ports": portsMap(containerInfo.ports),
		}
	}
	return nil
}

func (cb *Connbeat) exportServerConnection(s connections.ServerConnection, localIPs mapset.Set, containerInfo *ContainerInfo) error {
	event := common.MapStr{
		"@timestamp":    common.Time(time.Now()),
		"type":          "connbeat",
		"local_port":    s.LocalPort,
		"local_process": processAsMap(s.Process),
		"beat": common.MapStr{
			"local_ips": localIPs.ToSlice(),
		},
	}

	if containerInfo != nil {
		event["container"] = toMap(containerInfo)
	}

	cb.events.PublishEvent(event)

	return nil
}

func (cb *Connbeat) exportFullConnection(c connections.FullConnection, localIPs mapset.Set, containerInfo *ContainerInfo) error {
	event := common.MapStr{
		"@timestamp":    common.Time(time.Now()),
		"type":          "connbeat",
		"local_ip":      c.LocalIP,
		"local_port":    c.LocalPort,
		"remote_ip":     c.RemoteIP,
		"remote_port":   c.RemotePort,
		"local_process": processAsMap(c.Process),
		"beat": common.MapStr{
			"local_ips": localIPs.ToSlice(),
		},
	}

	if containerInfo != nil {
		event["container"] = toMap(containerInfo)
	}

	cb.events.PublishEvent(event)

	return nil
}

func update(infos map[string]ContainerInfo, socketContainerInfo *sockets.ContainerInfo, ip string) *ContainerInfo {
	if socketContainerInfo == nil {
		return nil
	}

	info, found := infos[socketContainerInfo.ID]
	if !found {
		localIPs := mapset.NewSet()
		info = ContainerInfo{
			socketContainerInfo.ID,
			socketContainerInfo.Name,
			socketContainerInfo.Image,
			localIPs,
			socketContainerInfo.DockerEnvironment,
			socketContainerInfo.DockerLabels,
			socketContainerInfo.Ports,
			socketContainerInfo.DockerhostHostname,
			socketContainerInfo.DockerhostIP,
		}
		infos[socketContainerInfo.ID] = info
	}
	if shouldBeRecorded(ip) {
		info.localIPs.Add(ip)
	}
	return &info
}

func (cb *Connbeat) Pipe(connectionListener <-chan connections.FullConnection, serverConnectionListener <-chan connections.ServerConnection) error {
	var err error

	localIPs := mapset.NewSet()
	containerInfo := make(map[string]ContainerInfo)
	container := (*ContainerInfo)(nil)

	for {
		select {
		case <-cb.done:
			return nil
		case c := <-connectionListener:
			container = update(containerInfo, c.Container, c.LocalIP)
			if container == nil && shouldBeRecorded(c.LocalIP) {
				localIPs.Add(c.LocalIP)
			}

			err = cb.exportFullConnection(c, localIPs, container)
			if err != nil {
				return err
			}
		case s := <-serverConnectionListener:
			container = update(containerInfo, s.Container, s.LocalIP)
			if container == nil && shouldBeRecorded(s.LocalIP) {
				localIPs.Add(s.LocalIP)
			}

			err = cb.exportServerConnection(s, localIPs, container)
			if err != nil {
				return err
			}
		}
	}
}

func isLocal(ip string) bool {
	return ip == "127.0.0.1" || ip == "::1"
}

func isWildcard(ip string) bool {
	return ip == "0.0.0.0" || ip == "::"
}

func shouldBeRecorded(ip string) bool {
	return !isWildcard(ip) && !isLocal(ip)
}

func (cb *Connbeat) Run(b *beat.Beat) error {
	connectionListener, serverConnectionListener, err := connections.Listen(
		cb.ConnConfig.Connbeat.ExposeProcessInfo, cb.ConnConfig.Connbeat.ExposeCmdline, cb.ConnConfig.Connbeat.ExposeEnviron,
		cb.ConnConfig.Connbeat.LocalConnectionsEnabled, cb.ConnConfig.Connbeat.DockerEnabled, cb.ConnConfig.Connbeat.TcpDiagEnabled,
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
