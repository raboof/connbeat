package beater

import (
	"time"
)

type ConnConfig struct {
	ExposeProcessInfo     bool          `config:"expose_process_info"`
	ExposeCmdline         bool          `config:"expose_cmdline"`
	ExposeEnviron         bool          `config:"expose_environ"`
	ConnectionAggregation time.Duration `config:"aggregation"`
	TcpDiagEnabled        bool          `config:"enable_tcp_diag"`
	PollInterval          time.Duration `config:"poll_interval"`
}

type ConfigSettings struct {
	Connbeat ConnConfig `config:"connbeat"`
}

func (c *ConnConfig) Validate() error {
	return nil
}

var (
	defaultConfig = ConnConfig{
		ExposeProcessInfo:     true,
		ExposeCmdline:         true,
		ExposeEnviron:         false,
		ConnectionAggregation: 30 * time.Second,
		TcpDiagEnabled:        false,
		PollInterval:          2 * time.Second,
	}
)
