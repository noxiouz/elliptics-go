package elliptics

import (
	"os"
	"testing"

	"time"
)

const REMOTE_ENV_PARAM = `ELLIPTICS_REMOTE`

var REMOTE string = os.Getenv(REMOTE_ENV_PARAM)

func TestLoggerAndNode(t *testing.T) {
	// Test creation
	EllLog, err := NewFileLogger("/tmp/elliptics.log", DEBUG)
	if err != nil {
		t.Fatalf("Unable to create logger %s", err)
	}
	defer EllLog.Free()
	if EllLog.GetLevel() != DEBUG {
		t.Error("Wrong loglevel")
	}
	EllLog.Log(INFO, "started: %v, level: %d", time.Now(), INFO)

	if len(REMOTE) == 0 {
		t.Skipf("Skip this test. Set %s env variable", REMOTE_ENV_PARAM)
	}
	node, err := NewNode(EllLog)
	if err != nil {
		t.Fatalf("Can't create node: %s", err)
	}
	defer node.Free()

	node.SetTimeouts(100, 1000)

	if err = node.AddRemote(REMOTE); err != nil {
		t.Fatalf("Failed to add remote %s: %s", REMOTE, err)
	}
}
