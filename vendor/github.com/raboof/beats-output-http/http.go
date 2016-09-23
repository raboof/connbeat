package http

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/common/op"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/outputs/mode"
	"github.com/elastic/beats/libbeat/outputs/mode/modeutil"
	"github.com/elastic/beats/libbeat/outputs/transport"
)

type httpOutput struct {
	mode mode.ConnectionMode
}

func init() {
	outputs.RegisterOutputPlugin("http", New)
}

var (
	debugf = logp.MakeDebug("http")
)

var (
	// ErrNotConnected indicates failure due to client having no valid connection
	ErrNotConnected = errors.New("not connected")

	// ErrJSONEncodeFailed indicates encoding failures
	ErrJSONEncodeFailed = errors.New("json encode failed")
)

func New(beatName string, cfg *common.Config, _ int) (outputs.Outputer, error) {
	output := &httpOutput{}
	err := output.init(cfg)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (out *httpOutput) init(cfg *common.Config) error {
	config := defaultConfig
	logp.Info("Initializing HTTP output")
	if err := cfg.Unpack(&config); err != nil {
		return err
	}

	tlsConfig, err := outputs.LoadTLSConfig(config.TLS)
	if err != nil {
		return err
	}

	clients, err := modeutil.MakeClients(cfg, makeClientFactory(tlsConfig, &config, out))
	if err != nil {
		return err
	}

	maxRetries := config.MaxRetries
	maxAttempts := maxRetries + 1 // maximum number of send attempts (-1 = infinite)
	if maxRetries < 0 {
		maxAttempts = 0
	}

	var waitRetry = time.Duration(1) * time.Second
	var maxWaitRetry = time.Duration(60) * time.Second

	loadBalance := config.LoadBalance
	m, err := modeutil.NewConnectionMode(clients, modeutil.Settings{
		Failover:     !loadBalance,
		MaxAttempts:  maxAttempts,
		Timeout:      config.Timeout,
		WaitRetry:    waitRetry,
		MaxWaitRetry: maxWaitRetry,
	})
	if err != nil {
		return err
	}

	out.mode = m

	return nil
}

func makeClientFactory(
	tls *transport.TLSConfig,
	config *httpConfig,
	out *httpOutput,
) func(string) (mode.ProtocolClient, error) {
	logp.Info("Making client factory")
	return func(host string) (mode.ProtocolClient, error) {
		logp.Info("Making client for host" + host)
		hostURL, err := getURL(config.Protocol, 80, config.Path, host)
		if err != nil {
			logp.Err("Invalid host param set: %s, Error: %v", host, err)
			return nil, err
		}

		var proxyURL *url.URL
		if config.ProxyURL != "" {
			proxyURL, err = parseProxyURL(config.ProxyURL)
			if err != nil {
				return nil, err
			}

			logp.Info("Using proxy URL: %s", proxyURL)
		}

		params := config.Params
		if len(params) == 0 {
			params = nil
		}

		return NewClient(ClientSettings{
			URL:              hostURL,
			Proxy:            proxyURL,
			TLS:              tls,
			Username:         config.Username,
			Password:         config.Password,
			Parameters:       params,
			Timeout:          config.Timeout,
			CompressionLevel: config.CompressionLevel,
		})
	}
}

func (out *httpOutput) Close() error {
	return nil
}

func (out *httpOutput) PublishEvent(
	trans op.Signaler,
	opts outputs.Options,
	data outputs.Data,
) error {
	return out.mode.PublishEvent(trans, opts, data)
}

func (out *httpOutput) BulkPublish(
	trans op.Signaler,
	opts outputs.Options,
	data []outputs.Data,
) error {
	return out.mode.PublishEvents(trans, opts, data)
}

func parseProxyURL(raw string) (*url.URL, error) {
	url, err := url.Parse(raw)
	if err == nil && strings.HasPrefix(url.Scheme, "http") {
		return url, err
	}

	// Proxy was bogus. Try prepending "http://" to it and
	// see if that parses correctly.
	return url.Parse("http://" + raw)
}
