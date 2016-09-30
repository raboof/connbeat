package http

import (
	"time"

	"github.com/elastic/beats/libbeat/outputs"
)

type httpConfig struct {
	Protocol         string             `config:"protocol"`
	Path             string             `config:"path"`
	Params           map[string]string  `config:"parameters"`
	Username         string             `config:"username"`
	Password         string             `config:"password"`
	ProxyURL         string             `config:"proxy_url"`
	LoadBalance      bool               `config:"loadbalance"`
	CompressionLevel int                `config:"compression_level" validate:"min=0, max=9"`
	TLS              *outputs.TLSConfig `config:"tls"`
	MaxRetries       int                `config:"max_retries"`
	Timeout          time.Duration      `config:"timeout"`
}

var (
	defaultConfig = httpConfig{
		Protocol:         "",
		Path:             "",
		ProxyURL:         "",
		Username:         "",
		Password:         "",
		Timeout:          90 * time.Second,
		CompressionLevel: 0,
		TLS:              nil,
		MaxRetries:       3,
		LoadBalance:      true,
	}
)

func (c *httpConfig) Validate() error {
	if c.ProxyURL != "" {
		if _, err := parseProxyURL(c.ProxyURL); err != nil {
			return err
		}
	}

	return nil
}
