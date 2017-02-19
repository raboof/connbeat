package beater

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/deckarep/golang-set"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/raboof/connbeat/connections"
	"github.com/raboof/connbeat/processes"
	"github.com/raboof/connbeat/sockets"

	"github.com/stvp/assert"
)

type TestClient struct {
	evs chan common.MapStr
}

var (
	httpd = processes.UnixProcess{
		Binary:  "httpd",
		Cmdline: "/bin/httpd",
		Environ: "",
	}
	curl = processes.UnixProcess{
		Binary:  "curl",
		Cmdline: "/usr/bin/curl http://www.nu.nl",
		Environ: "",
	}
)

func (TestClient) Close() error { return nil }
func (tc TestClient) PublishEvents(events []common.MapStr, opts ...publisher.ClientOption) bool {
	for _, event := range events {
		tc.evs <- event
	}
	return true
}
func (tc TestClient) PublishEvent(event common.MapStr, opts ...publisher.ClientOption) bool {
	return tc.PublishEvents([]common.MapStr{event}, opts...)
}

func TestLocalIps(t *testing.T) {
	beater := &Connbeat{}

	fullConnections, serverConnections := make(chan connections.FullConnection), make(chan connections.ServerConnection)

	client := TestClient{
		evs: make(chan common.MapStr),
	}

	beater.events = client
	beater.done = make(chan struct{})

	go beater.Pipe(fullConnections, serverConnections)
	serverConnections <- connections.ServerConnection{"12.34.6.2", 80, &httpd, nil}
	_ = <-client.evs

	serverConnections <- connections.ServerConnection{"127.0.0.1", 80, &httpd, nil}
	_ = <-client.evs

	fullConnections <- connections.FullConnection{connections.LocalConnection{"43.12.1.32", 22, &curl, nil}, "43.23.2.4", 5113}
	evt := <-client.evs
	ips, err := evt.GetValue("beat.local_ips")
	if err != nil {
		fmt.Println(evt)
		t.FailNow()
	}

	expectElements(t, []string{"12.34.6.2", "43.12.1.32"}, ips.([]interface{}))
}

func TestNoContainerInfo(t *testing.T) {
	beater := &Connbeat{}

	fullConnections, serverConnections := make(chan connections.FullConnection), make(chan connections.ServerConnection)

	client := TestClient{
		evs: make(chan common.MapStr),
	}

	beater.events = client
	beater.done = make(chan struct{})

	go beater.Pipe(fullConnections, serverConnections)
	serverConnections <- connections.ServerConnection{"12.34.6.2", 80, &httpd, nil}
	evt := <-client.evs

	container, present := evt["container"]
	assert.False(t, present, "There should be no container field in the event")
	assert.Nil(t, container, "There should be no container field in the event")
}

func TestMapContainerInfoWithoutHostIp(t *testing.T) {
	containerInfo := &ContainerInfo{
		id:                 "7786521dc8c9",
		localIPs:           mapset.NewSet(),
		environment:        nil,
		dockerHostHostname: "yinka",
		dockerHostIP:       nil}
	json := toMap(containerInfo)
	ips, err := json.GetValue("docker_host.ips")
	if err != nil {
		t.Fatal("Failed to get docker_host.ips from event", json)
	} else {
		assert.Equal(t, 0, reflect.ValueOf(ips).Len(), "Expected empty list of container host ips")
	}
}

func TestMapContainerInfoWithHostIp(t *testing.T) {
	containerInfo := &ContainerInfo{
		id:                 "7786521dc8c9",
		localIPs:           mapset.NewSet(),
		environment:        nil,
		dockerHostHostname: "yinka",
		dockerHostIP:       net.IP("127.0.0.1")}
	json := toMap(containerInfo)
	ips, err := json.GetValue("docker_host.ips")
	if err != nil {
		t.Fatal("Failed to get docker_host.ips from event", json)
	} else {
		assert.Equal(t, 1, reflect.ValueOf(ips).Len(), "Expected list with one container host ip")
		assert.Equal(t, net.IP("127.0.0.1"), reflect.ValueOf(ips).Index(0).Interface(), "Expected container host ips")
	}
}

func TestContainerInformation(t *testing.T) {
	beater := &Connbeat{}

	fullConnections, serverConnections := make(chan connections.FullConnection), make(chan connections.ServerConnection)

	client := TestClient{
		evs: make(chan common.MapStr),
	}

	beater.events = client
	beater.done = make(chan struct{})

	dockerLabels := make(map[string]string)
	dockerLabels["test.test.1"] = "example"
	dockerLabels["example"] = "test@#$%&"

	go beater.Pipe(fullConnections, serverConnections)
	serverConnections <- connections.ServerConnection{"12.34.6.2", 80, &httpd, &sockets.ContainerInfo{
		ID:                 "7786521dc8c9",
		DockerhostHostname: "yinka",
		DockerhostIP:       nil}}
	_ = <-client.evs

	fullConnections <- connections.FullConnection{connections.LocalConnection{"43.12.1.32", 22, &curl, &sockets.ContainerInfo{
		ID:                 "785073e68b72",
		DockerEnvironment:  nil,
		DockerLabels:       dockerLabels,
		DockerhostHostname: "yinka",
		DockerhostIP:       nil}}, "43.23.2.4", 5113}
	evt := <-client.evs
	ips, err := evt.GetValue("beat.local_ips")
	if err != nil {
		fmt.Println(evt)
		t.FailNow()
	}

	expectElements(t, []string{}, ips.([]interface{}))

	containerIps, err := evt.GetValue("container.local_ips")
	if err != nil {
		t.Fatal(err)
	}
	expectElements(t, []string{"43.12.1.32"}, containerIps.([]interface{}))

	labels, err := evt.GetValue("container.labels")
	if err != nil {
		t.Fatal(err)
	}

	expectMap(t, dockerLabels, labels.(common.MapStr))
}

func TestNoContainerInformationLeakage(t *testing.T) {
	beater := &Connbeat{}

	fullConnections, serverConnections := make(chan connections.FullConnection), make(chan connections.ServerConnection)

	client := TestClient{
		evs: make(chan common.MapStr),
	}

	beater.events = client
	beater.done = make(chan struct{})

	go beater.Pipe(fullConnections, serverConnections)
	serverConnections <- connections.ServerConnection{"12.34.6.2", 80, &httpd, &sockets.ContainerInfo{
		ID:                 "7786521dc8c9",
		DockerhostHostname: "yinka",
		DockerhostIP:       nil}}
	_ = <-client.evs

	fullConnections <- connections.FullConnection{connections.LocalConnection{"43.12.1.32", 22, &curl, nil}, "43.23.2.4", 5113}
	evt := <-client.evs

	container, _ := evt.GetValue("container")
	assert.Nil(t, container, "Container information should not leak into the second event")
}

func expectMap(t *testing.T, expected map[string]string, actual common.MapStr) {
	assert.Equal(t, len(actual), len(expected), "should have the expected number of elements")
	for expectedKey, expectedValue := range expected {
		//Using bracket notation instead of getValue, to correctly handle keys that have dots.
		actualValue := actual[expectedKey]
		assert.Equal(t, expectedValue, actualValue, "values should be equal")
	}
}

func expectElements(t *testing.T, expected []string, actual []interface{}) {
	assert.Equal(t, len(actual), len(expected), "should have the expected number of elements")
	for _, expectation := range expected {
		expectElement(t, expectation, actual)
	}
}

func expectElement(t *testing.T, expected string, actual []interface{}) {
	for _, found := range actual {
		if expected == found {
			return
		}
	}
	fmt.Printf("Expected but not found: %s\n", expected)
	t.FailNow()
}
