package elliptics

import (
	"log"
	"os"
	"testing"
)

const REMOTE_ENV_PARAM = `ELLIPTICS_REMOTE`

var REMOTE string = os.Getenv(REMOTE_ENV_PARAM)

func TestSession(t *testing.T) {
	var (
		sessionGroups  = []int32{1, 2, 100, 505}
		sessionTraceID = uint64(99999)
	)

	const (
		sessionNamespace = "sessionNamespace"
		sessionTimeout   = 5
	)

	l := log.New(os.Stderr, "TEST", log.Ltime)

	node, err := NewNode(l, "info")
	if err != nil {
		t.Fatalf("NewNode: unexpected error %s", err)
	}

	session, err := NewSession(node)
	if err != nil {
		t.Fatalf("NewSession: unexpected error %s", err)
	}

	session.SetGroups(sessionGroups)

	if gotGroups := session.GetGroups(); len(gotGroups) != len(sessionGroups) {
		t.Errorf("SetGroups & GetGroups: invalid groups number. expected %d, got %d",
			len(sessionGroups), len(gotGroups))
	}

	session.SetNamespace(sessionNamespace)
	session.SetTimeout(sessionTimeout)
	session.SetTraceID(sessionTraceID)
}
