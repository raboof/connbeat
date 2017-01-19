package beater

import (
	"errors"
	"time"

	"github.com/elastic/beats/libbeat/logp"
)

type ConnConfig struct {
	ExposeProcessInfo       bool          `config:"expose_process_info"`
	ExposeCmdline           bool          `config:"expose_cmdline"`
	ExposeEnviron           bool          `config:"expose_environ"`
	ConnectionAggregation   time.Duration `config:"aggregation"`
	LocalConnectionsEnabled bool          `config:"enable_local_connections"`
	DockerEnabled           bool          `config:"enable_docker"`
	DockerEnvironment       []string      `config:"docker_environment"`
	TcpDiagEnabled          bool          `config:"enable_tcp_diag"`
	PollInterval            time.Duration `config:"poll_interval"`
}

type ConfigSettings struct {
	Connbeat ConnConfig `config:"connbeat"`
}

func (c *ConnConfig) Validate() error {
	if c.DockerEnabled && c.TcpDiagEnabled {
		if c.LocalConnectionsEnabled {
			logp.Debug("connbeat", "tcp_diag enabled for local processes but not for docker")
		} else {
			return errors.New("tcp_diag is not currently supported when monitoring docker instances")
		}
	}
	return nil
}

var (
	defaultConfig = ConnConfig{
		ExposeProcessInfo:       true,
		ExposeCmdline:           true,
		ExposeEnviron:           false,
		ConnectionAggregation:   30 * time.Second,
		LocalConnectionsEnabled: true,
		DockerEnabled:           false,
		DockerEnvironment:       nil,
		TcpDiagEnabled:          false,
		PollInterval:            2 * time.Second,
	}
)
