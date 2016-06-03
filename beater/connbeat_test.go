package beater

import (
	"fmt"
	"testing"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/publisher"
)

type TestClient struct {
	evs chan common.MapStr
}

func (TestClient) Close() error { return nil }
func (tc TestClient) PublishEvents(events []common.MapStr, opts ...publisher.ClientOption) bool {
	for _, event := range events {
		tc.evs <- event
	}
	// fmt.Printf("Publish! sink is now %d\n", len(*tc.sink))
	return true
}
func (tc TestClient) PublishEvent(event common.MapStr, opts ...publisher.ClientOption) bool {
	return tc.PublishEvents([]common.MapStr{event}, opts...)
}

func TestLocalIps(t *testing.T) {
	beater := New()

	connections, serverConnections := make(chan Connection), make(chan ServerConnection)

	client := TestClient{
		evs: make(chan common.MapStr),
	}

	beater.events = client
	beater.done = make(chan struct{})

	go beater.Pipe(connections, serverConnections)
	serverConnections <- ServerConnection{"12.34.6.2", 80, "httpd"}
	_ = <-client.evs

	connections <- Connection{"43.12.1.32", 22, "43.23.2.4", 5113, "asdf"}
	evt := <-client.evs
	ips, err := evt.GetValue("beat.local_ips")
	if err != nil {
		fmt.Println(evt)
		t.FailNow()
	}

	expectElements(t, ips.([]interface{}), []string{"12.34.6.2", "43.12.1.32"})
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
