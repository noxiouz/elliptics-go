package elliptics

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

const REMOTE_ENV_PARAM = `ELLIPTICS_REMOTE`

var REMOTE string = os.Getenv(REMOTE_ENV_PARAM)

func TestSession(t *testing.T) {
	var (
		sessionGroups  = []int32{1, 2, 100, 505}
		sessionTraceID = TraceID(99999)

		sessionCflags  = DNET_FLAGS_NOCACHE
		sessionIOflags = DNET_IO_FLAGS_NOCSUM
	)

	const (
		sessionNamespace = "sessionNamespace"
		sessionTimeout   = 5
	)

	l := log.New(ioutil.Discard, "TEST", log.Ltime)

	node, err := NewNode(l, "error")
	if err != nil {
		t.Fatalf("NewNode: unexpected error %s", err)
	}

	session, err := NewSession(node)
	if err != nil {
		t.Fatalf("NewSession: unexpected error %s", err)
	}

	session.SetGroups(sessionGroups)

	if gotGroups := session.GetGroups(); len(gotGroups) != len(sessionGroups) {
		t.Errorf("SetGroups & GetGroups: invalid groups number. Expected %d, got %d",
			len(sessionGroups), len(gotGroups))
	}

	session.SetNamespace(sessionNamespace)

	session.SetTimeout(sessionTimeout)
	if gotTimeout := session.GetTimeout(); gotTimeout != sessionTimeout {
		t.Errorf("Set/GetTimeout: invalid timeout value. Expected %d, got %d",
			sessionTimeout, gotTimeout)
	}

	session.SetTraceID(sessionTraceID)
	if gotTraceId := session.GetTraceID(); gotTraceId != sessionTraceID {
		t.Errorf("Set/GetTraceID: invalid timeout value. Expected %d, got %d",
			sessionTraceID, gotTraceId)
	}

	session.SetCflags(sessionCflags)
	if gotCflags := session.GetCflags(); gotCflags != sessionCflags {
		t.Errorf("Set/GetCflags: invalid timeout value. Expected %d, got %d",
			sessionCflags, gotCflags)
	}

	session.SetIOflags(sessionIOflags)
	if gotIOflags := session.GetIOflags(); gotIOflags != sessionIOflags {
		t.Errorf("Set/GetIOflags: invalid timeout value. Expected %d, got %d",
			sessionIOflags, gotIOflags)
	}
}
