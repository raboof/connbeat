package beater

import (
  "testing"

  "github.com/stvp/assert"
)

func TestErrorOnConflictingOptions(t *testing.T) {
  config := ConnConfig{
		DockerEnabled:         true,
		TcpDiagEnabled:        true,
	}
  err := config.Validate()
  assert.NotNil(t, err, "should produce an error when enabling both tcp_diag and docker")
}

func TestNoErrorForDefaultConfig(t *testing.T) {
  err := defaultConfig.Validate()
  assert.Nil(t, err, "should not produce an error for the default options ")
}
