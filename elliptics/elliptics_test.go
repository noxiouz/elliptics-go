package elliptics

import (
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	var (
		sessionGroups  = []uint32{1, 2, 100, 505}
		sessionTraceID = TraceID(99999)

		sessionCflags  = DNET_FLAGS_NOCACHE
		sessionIOflags = DNET_IO_FLAGS_NOCSUM
	)

	const (
		sessionNamespace = "sessionNamespace"
		sessionTimeout   = 5
	)

	node, err := NewNode("/dev/stderr", "error")
	if err != nil {
		t.Fatalf("NewNode: unexpected error %s", err)
	}

	defer func() {
		time.Sleep(1 * time.Second)
		node.Free()
	}()

	session, err := NewSession(node)
	if err != nil {
		t.Fatalf("NewSession: unexpected error %s", err)
	}
	defer session.Delete()

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

	dnetStat := session.DnetStat()
	t.Log(dnetStat)
	defer session.GetRoutes(dnetStat)
	t.Log(dnetStat.StatData())
}

func TestFull(t *testing.T) {
	/* Preparing */

	const (
		FULL_TEST_REMOTES = `FULL_TEST_REMOTES`
		FULL_TEST_GROUPS  = `FULL_TEST_GROUPS`

		TEST_BLOB = `MY_TEST_BLOB_WITH_DUMMY_DATA`
	)
	var (
		testRemotes string = os.Getenv(FULL_TEST_REMOTES)
		testGroups  string = os.Getenv(FULL_TEST_GROUPS)

		testKey        string    = fmt.Sprintf("testkey-%d", time.Now().Unix())
		testNamespace  string    = fmt.Sprintf("testnamespace-%d", time.Now().Unix())
		testBlobReader io.Reader = strings.NewReader(TEST_BLOB)
	)

	if testRemotes == "" || testGroups == "" {
		t.Log(testRemotes, testGroups)
		t.Skipf(`TestFull: Skipped as remotes and groups aren't specified.
Setup env variables. Example: export %s="localhost:1025:2" && export %s="1,2,3"`,
			FULL_TEST_REMOTES, FULL_TEST_GROUPS)
	}

	Remotes := strings.Split(testRemotes, ",")

	raw_groups := strings.Split(testGroups, ",")
	Groups := make([]uint32, len(raw_groups))
	for i, groups := range raw_groups {
		gr, err := strconv.Atoi(groups)
		if err != nil {
			t.Fatalf("TestFull: invalid group number %s", err)
		} else {
			Groups[i] = uint32(gr)
		}
	}

	indexes_names := []string{"A", "B"}
	bad_indexes_names := []string{"Y", "Z"}
	indexes := make(map[string]string)
	reverse_indexes := make(map[string]string)
	for _, index_name := range indexes_names {
		value := fmt.Sprintf("extended_value_%s", index_name)
		indexes[index_name] = value
		reverse_indexes[value] = index_name
	}

	extended_indexes := make(map[string]string)
	extended_reverse_indexes := make(map[string]string)
	extended_indexes_names := append(indexes_names, bad_indexes_names...)
	for _, index_name := range extended_indexes_names {
		value := fmt.Sprintf("value%s", index_name)
		extended_indexes[index_name] = value
		extended_reverse_indexes[value] = index_name
	}
	/*  End of the preparing */

	//Create Node
	node, err := NewNode("/dev/stderr", "error")
	if err != nil {
		t.Fatalf("unexpected error %s during NewNode", err)
	}

	t.Logf("Adding remotes %s and groups %v", Remotes, Groups)
	if err = node.AddRemotes(Remotes); err != nil {
		t.Fatalf("unexpected error %s during AddRemotes", err)
	}

	defer func() {
		time.Sleep(1 * time.Second)
		node.Free()
	}()

	session, err := NewSession(node)
	if err != nil {
		t.Fatalf("unexpected error %s during NewSession", err)
	}
	session.SetGroups(Groups)
	session.SetNamespace(testNamespace)

	for res := range session.WriteData(testKey, testBlobReader, 0, 0) {
		if err := res.Error(); err != nil {
			t.Fatalf("lookup result error: %s", err)
		}
	}

	key, _ := NewKey(testKey)
	for res := range session.Lookup(key) {
		dnet_add := res.Addr()

		if len(dnet_add.String()) == 0 {
			t.Fatalf("session.Lookup error: invalid String()")
		}

		if err := res.Error(); err != nil {
			t.Fatalf("session.Lookup error: %s", err)
		}

	}

	for res := range session.ParallelLookup(testKey) {
		if err := res.Error(); err != nil {
			t.Fatalf("session.Lookup error: %s", err)
		}
	}

	for res := range session.ReadData(testKey, 1, 1) {
		if err := res.Error(); err != nil {
			t.Fatalf("session.ReadData error: %s", err)
		}

		if res_size := len(res.Data()); res_size != 1 {
			t.Fatalf("session.ReadData: wrong response size. Expected 1, got %d", res_size)
		}

		if res.Data()[0] != TEST_BLOB[1] {
			t.Fatalf("session.ReadData: wrong response content. Expected %v, got %v", TEST_BLOB[1], res.Data()[0])
		}
	}

	w := httptest.NewRecorder()
	size := uint64(10)
	if err := session.StreamHTTP(testKey, 0, size, w); err != nil {
		t.Fatalf("session.StreamHTTP. Unexpected error %s", err)
	}
	if uint64(w.Body.Len()) != size {
		t.Errorf("session.StreamHTTP. Invalid Body length")
	}

	for res := range session.SetIndexes(testKey, indexes) {
		if err := res.Error(); err != nil {
			t.Fatalf("session.SetIndexes error: %s", err)
		}
	}

	var i int = 0
	for res := range session.ListIndexes(testKey) {
		if err := res.Error(); err != nil {
			t.Fatalf("session.ListIndexes error: %s", err)
		}

		i += 1
		index_item_name := res.Data
		if _, ok := reverse_indexes[index_item_name]; !ok {
			t.Fatalf("session.ListIndexes error: unset index is found %s", index_item_name)
		}
	}

	if i != len(indexes_names) {
		t.Fatalf("session.ListIndexes error: invalid total numbes of indxes. Expected %d, got %d", len(indexes_names), i)
	}

	// FindAll. Must be empty
	for res := range session.FindAllIndexes(append(indexes_names, bad_indexes_names...)) {
		t.Fatalf("Result must empty, but got %v", res)
	}

	// FindAny
	i = 0
	for res := range session.FindAnyIndexes(append(bad_indexes_names, indexes_names[0])) {
		i += 1
		t.Logf("%s", res.Data()[0].Data)
	}

	if i != 1 {
		t.Fatalf("session.FindAnyIndexes error: invalid total numbes of indxes. Expected 1, got %d", i)
	}
	//=====================

	// Case: Update indexes, remove old ones, and list
	for res := range session.UpdateIndexes(testKey, extended_indexes) {
		if err := res.Error(); err != nil {
			t.Fatalf("session.UpdateIndexes error: %s", err)
		}
	}

	for res := range session.RemoveIndexes(testKey, indexes_names) {
		if err := res.Error(); err != nil {
			t.Fatalf("session.RemoveIndexes error: %s", err)
		}
	}

	i = 0
	for res := range session.ListIndexes(testKey) {
		if err := res.Error(); err != nil {
			t.Fatalf("session.ListIndexes error: %s", err)
		}

		i += 1
		index_item_name := res.Data
		if _, ok := reverse_indexes[index_item_name]; ok {
			t.Fatalf("session.ListIndexes error: removed index is found %s", index_item_name)
		}
	}

	if i != len(bad_indexes_names) {
		t.Fatalf("session.ListIndexes error: invalid total numbes of indxes. Expected %d, got %d", len(bad_indexes_names), i)
	}
	//================================================

	// Remove
	for res := range session.Remove(testKey) {
		if err := res.Error(); err != nil {
			t.Fatalf("session.Remove error: %s", err)
		}
	}

	// Lookup after Remove
	for res := range session.Lookup(key) {
		if err := res.Error(); err == nil {
			t.Fatalf("session.Lookup. Expected error, but got nil")
		}
	}

}
