package beater

import (
	"time"
)

type ConnConfig struct {
	ConnectionAggregation time.Duration `config:"aggregation"`
}

type ConfigSettings struct {
	Connbeat ConnConfig `config:"connbeat"`
}

var (
	defaultConfig = ConnConfig{
		ConnectionAggregation: 30 * time.Second,
	}
)
