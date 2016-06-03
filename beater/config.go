package beater

import (
	"time"
)

type ConnConfig struct {
	ExposeCmdline         bool          `config:"expose_cmdline"`
	ExposeEnviron         bool          `config:"expose_environ"`
	ConnectionAggregation time.Duration `config:"aggregation"`
}

type ConfigSettings struct {
	Connbeat ConnConfig `config:"connbeat"`
}

func (c *ConnConfig) Validate() error {
	return nil
}

var (
	defaultConfig = ConnConfig{
		ExposeCmdline:         true,
		ExposeEnviron:         false,
		ConnectionAggregation: 30 * time.Second,
	}
)
