package docker

import (
	"net"
	"os"
	"testing"

	// "golang.org/x/net/context"

	"github.com/fsouza/go-dockerclient"
	docker_testing "github.com/fsouza/go-dockerclient/testing"

	"github.com/stvp/assert"
)

func TestContainerMetadata(t *testing.T) {
	server, err := docker_testing.NewServer("127.0.0.1:0", nil, nil)
	defer server.Stop()
	assert.Nil(t, err, "creating server")

	client, err := docker.NewClient(server.URL())
	assert.Nil(t, err, "creating client")

	err = client.PullImage(docker.PullImageOptions{
		Repository: "asdf",
		Tag:        "latest",
	}, docker.AuthConfiguration{
		Username: "user",
		Password: "pass",
	})
	assert.Nil(t, err, "introducing image")

	_, err = client.CreateContainer(docker.CreateContainerOptions{
		Name:   "myasdf",
		Config: &docker.Config{Image: "asdf:latest", AttachStdout: true, AttachStdin: true},
	})
	assert.Nil(t, err, "creating container")
	err = client.StartContainer("myasdf", &docker.HostConfig{})
	assert.Nil(t, err, "starting container")

	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	assert.Nil(t, err, "listing containers")

	assert.Equal(t, 1, len(containers), "number of containers")
	assert.Equal(t, containers[0].Names, []string{"/myasdf"}, "should use the name used when creating the container")
	assert.Equal(t, containers[0].Image, "asdf:latest", "should use the image used when creating the container")

}

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
	assert.Equal(t, net.ParseIP("11.11.11.11"), poller.dockerhostIP, "should get IP address from environment")
}
