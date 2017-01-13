package beater

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/deckarep/golang-set"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"

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

	connections, serverConnections := make(chan Connection), make(chan ServerConnection)

	client := TestClient{
		evs: make(chan common.MapStr),
	}

	beater.events = client
	beater.done = make(chan struct{})

	go beater.Pipe(connections, serverConnections)
	serverConnections <- ServerConnection{"12.34.6.2", 80, &httpd, nil}
	_ = <-client.evs

	connections <- Connection{"43.12.1.32", 22, "43.23.2.4", 5113, &curl, nil}
	evt := <-client.evs
	ips, err := evt.GetValue("beat.local_ips")
	if err != nil {
		fmt.Println(evt)
		t.FailNow()
	}

	expectElements(t, ips.([]interface{}), []string{"12.34.6.2", "43.12.1.32"})
}

func TestNoContainerInfo(t *testing.T) {
	beater := &Connbeat{}

	connections, serverConnections := make(chan Connection), make(chan ServerConnection)

	client := TestClient{
		evs: make(chan common.MapStr),
	}

	beater.events = client
	beater.done = make(chan struct{})

	go beater.Pipe(connections, serverConnections)
	serverConnections <- ServerConnection{"12.34.6.2", 80, &httpd, nil}
	evt := <-client.evs

	container, present := evt["container"]
	assert.False(t, present, "There should be no container field in the event")
	assert.Nil(t, container, "There should be no container field in the event")
}

func TestMapContainerInfoWithoutHostIp(t *testing.T) {
	containerInfo := &ContainerInfo{
		id:                 "7786521dc8c9",
		localIPs: mapset.NewSet(),
		environment: nil,
		dockerHostHostname: "yinka",
		dockerHostIP:       nil}
	json := toMap(containerInfo)
	ips, err := json.GetValue("docker_host.ips")
	if err != nil {
		t.Fatal("Failed to get docker_host.ips from event", json)
	} else {
		assert.Equal(t, 0, reflect.ValueOf(ips).Len(), "Expected empty list of container host ips");
	}
}

func TestMapContainerInfoWithHostIp(t *testing.T) {
	containerInfo := &ContainerInfo{
		id:                 "7786521dc8c9",
		localIPs: mapset.NewSet(),
		environment: nil,
		dockerHostHostname: "yinka",
		dockerHostIP:       net.IP("127.0.0.1")}
	json := toMap(containerInfo)
	ips, err := json.GetValue("docker_host.ips")
	if err != nil {
		t.Fatal("Failed to get docker_host.ips from event", json)
	} else {
		assert.Equal(t, 1, reflect.ValueOf(ips).Len(), "Expected empty list of container host ips");
		assert.Equal(t, net.IP("127.0.0.1"), reflect.ValueOf(ips).Index(0).Interface(), "Expected container host ips");
	}
}

func TestContainerInformation(t *testing.T) {
	beater := &Connbeat{}

	connections, serverConnections := make(chan Connection), make(chan ServerConnection)

	client := TestClient{
		evs: make(chan common.MapStr),
	}

	beater.events = client
	beater.done = make(chan struct{})

	go beater.Pipe(connections, serverConnections)
	serverConnections <- ServerConnection{"12.34.6.2", 80, &httpd, &sockets.ContainerInfo{
		ID:                 "7786521dc8c9",
		DockerhostHostname: "yinka",
		DockerhostIP:       nil}}
	_ = <-client.evs

	connections <- Connection{"43.12.1.32", 22, "43.23.2.4", 5113, &curl, &sockets.ContainerInfo{
		ID:                 "785073e68b72",
		DockerEnvironment:  nil,
		DockerhostHostname: "yinka",
		DockerhostIP:       nil}}
	evt := <-client.evs
	ips, err := evt.GetValue("beat.local_ips")
	if err != nil {
		fmt.Println(evt)
		t.FailNow()
	}

	expectElements(t, ips.([]interface{}), []string{"12.34.6.2", "43.12.1.32"})

	containerIps, err := evt.GetValue("container.local_ips")
	if err != nil {
		t.Fatal(err)
	}
	expectElements(t, containerIps.([]interface{}), []string{"43.12.1.32"})
}

func expectElements(t *testing.T, actual []interface{}, expected []string) {
	for _, expectation := range expected {
		expectElement(t, actual, expectation)
	}
}

func expectElement(t *testing.T, actual []interface{}, expected string) {
	for _, found := range actual {
		if expected == found {
			return
		}
	}
	fmt.Printf("Expected but not found: %s\n", expected)
	t.FailNow()
}
