package elliptics

import (
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func TestNode(t *testing.T) {
	const (
		suitableLogLevel = "info"
		invalidLogLevel  = "DEVNULLISHE"

		malformedAddress = "blabla:1025:22"
	)
	l := log.New(ioutil.Discard, "TEST", log.Ltime)

	if _, err := NewNode(l, invalidLogLevel); err == nil {
		t.Errorf("NewNode: error was expected, got nil")
	}

	node, err := NewNode(l, suitableLogLevel)
	if err != nil {
		t.Fatalf("NewNode: unexpected error %s", err)
	}

	node.SetTimeouts(5, 60)

	if err := node.AddRemote(malformedAddress); err == nil {
		t.Errorf("AddRemote: error was expected, but got nil")
	}

	malformedAddresses := []string{malformedAddress, malformedAddress}
	if err := node.AddRemotes(malformedAddresses); err == nil {
		t.Errorf("AddRemote: error was expected, but got nil")
	}

	if err := node.AddRemotes([]string{}); err == nil {
		t.Errorf("AddRemote: error was expected, but got nil")
	}

	time.Sleep(1 * time.Second)
	node.Free()
}
