package docker

import (
	"net"
	"os"
	"testing"

	"github.com/fsouza/go-dockerclient"
	docker_testing "github.com/fsouza/go-dockerclient/testing"

	"github.com/stvp/assert"
)

func TestHostnameFromEnvironment(t *testing.T) {
	os.Setenv("DOCKERHOST_HOSTNAME", "overridden.example")
	hostname, err := getDockerhostHostname(nil)
	assert.Nil(t, err, err)
	assert.Equal(t, "overridden.example", hostname, "should get hostname from environment")
}

func TestHostnameFromClient(t *testing.T) {
	os.Unsetenv("DOCKERHOST_HOSTNAME")
	server, err := docker_testing.NewServer("127.0.0.1:0", nil, nil)
	defer server.Stop()
	assert.Nil(t, err, "creating server")

	client, err := docker.NewClient(server.URL())
	assert.Nil(t, err, "creating client")

	hostname, err := getDockerhostHostname(client)
	assert.Nil(t, err, err)
	assert.Equal(t, "vagrant-ubuntu-trusty-64", hostname, "should get hostname from docker info")
}

func TestDockerhostIPFromEnvironment(t *testing.T) {
	os.Setenv("DOCKERHOST_IP", "11.11.11.11")
	server, err := docker_testing.NewServer("127.0.0.1:0", nil, nil)
	defer server.Stop()
	assert.Nil(t, err, "creating server")

	client, err := docker.NewClient(server.URL())
	assert.Nil(t, err, "creating client")

	poller, err := new(client, []string{})
	assert.Equal(t, net.ParseIP("11.11.11.11"), poller.dockerhostIP, "IP address from environment")
}
