package elliptics

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

func init() {
	Suite(&SessionSuite{})
}

type SessionSuite struct {
	NodeSuite
	session *Session
	groups  []uint32
}

func (s *SessionSuite) SetUpSuite(c *C) {
	const (
		testRemotesEnv = `TEST_REMOTES`
		testGroupsEnv  = `TEST_GROUPS`
	)
	var (
		testRemotes = os.Getenv(testRemotesEnv)
		testGroups  = os.Getenv(testGroupsEnv)
	)

	if testRemotes == "" || testGroups == "" {
		c.Log(testRemotes, testGroups)
		c.Skip(fmt.Sprintf(`TestFull: Skipped as remotes and groups aren't specified.
	Setup env variables. Example: export %s="localhost:1025:2" && export %s="1,2,3"`,
			testRemotesEnv, testGroupsEnv))
	}

	s.NodeSuite.SetUpTest(c)
	s.node.AddRemotes(strings.Split(testRemotes, ","))
	for _, group := range strings.Split(testGroups, ",") {
		gr, err := strconv.ParseUint(group, 10, 32)
		if err != nil {
			c.Fatalf("TestFull: invalid group number %v", err)
		}
		s.groups = append(s.groups, uint32(gr))
	}
}

func (s *SessionSuite) TearDownSuite(c *C) {
	s.NodeSuite.TearDownTest(c)
}

func (s *SessionSuite) SetUpTest(c *C) {
	session, err := NewSession(s.node)
	c.Assert(err, IsNil)

	s.session = session
}

func (s *SessionSuite) TearDownTest(c *C) {
	if s.session != nil {
		s.session.Delete()
	}
}

func (s *SessionSuite) TestGroups(c *C) {
	var sessionGroups = []uint32{1, 2, 100, 505}
	s.session.SetGroups(sessionGroups)
	c.Assert(s.session.GetGroups(), DeepEquals, sessionGroups)
}

func (s *SessionSuite) TestClone(c *C) {
	var sessionGroups = []uint32{1, 2, 100, 505}
	s.session.SetGroups(sessionGroups)
	s.session.SetCflags(DNET_FLAGS_NOCACHE)
	s.session.SetIOflags(DNET_IO_FLAGS_NOCSUM)

	session, err := CloneSession(s.session)
	c.Assert(err, IsNil)
	defer session.Delete()

	c.Assert(session.groups, DeepEquals, s.session.groups)
	c.Assert(session.GetCflags(), Equals, s.session.GetCflags())
	c.Assert(session.GetIOflags(), Equals, s.session.GetIOflags())
}

func (s *SessionSuite) TestTimeout(c *C) {
	const sessionTimeout = 5
	s.session.SetTimeout(sessionTimeout)
	c.Assert(s.session.GetTimeout(), Equals, sessionTimeout)
}

func (s *SessionSuite) TestNamespace(c *C) {
	const sessionNamespace = "sessionNamespace"
	s.session.SetNamespace(sessionNamespace)
}

func (s *SessionSuite) TestTraceID(c *C) {
	var sessionTraceID = TraceID(99999)
	s.session.SetTraceID(sessionTraceID)
	c.Assert(s.session.GetTraceID(), Equals, sessionTraceID)
}

func (s *SessionSuite) TestCFlags(c *C) {
	var sessionCflags = DNET_FLAGS_NOCACHE
	s.session.SetCflags(sessionCflags)
	c.Assert(s.session.GetCflags(), Equals, sessionCflags)
}

func (s *SessionSuite) TestIOFlags(c *C) {
	var sessionIOflags = DNET_IO_FLAGS_NOCSUM
	s.session.SetIOflags(sessionIOflags)
	c.Assert(s.session.GetIOflags(), Equals, sessionIOflags)
}

func (s *SessionSuite) TestSessionStat(c *C) {
	dnetStat := s.session.DnetStat()
	c.Log(dnetStat)
	defer s.session.GetRoutes(dnetStat)
	c.Log(dnetStat.StatData())
}

func (s *SessionSuite) TestTimestamp(c *C) {
	ts := time.Now()
	s.session.SetTimestamp(ts)
	c.Assert(s.session.GetTimestamp(), Equals, ts)
}

// TestReadWrite writes a key with a data, then read it
func (s *SessionSuite) TestWriteRead(c *C) {
	var (
		testBlob                 = `MY_TEST_BLOB_WITH_DUMMY_DATA`
		testKey                  = fmt.Sprintf("testkey-%d", time.Now().Unix())
		testNamespace            = fmt.Sprintf("testnamespace-%d", time.Now().Unix())
		testBlobReader io.Reader = strings.NewReader(testBlob)
	)

	s.session.SetGroups(s.groups)
	s.session.SetNamespace(testNamespace)

	for res := range s.session.WriteData(testKey, testBlobReader, 0, 0) {
		c.Assert(res.Error(), IsNil)
	}

	// No overflow must be there
	// TODO: addd Assert
	offset, size := uint64(1), uint64(2)
	for res := range s.session.ReadData(testKey, offset, size) {
		// No read error
		c.Assert(res.Error(), IsNil)
		// Read exactly size
		c.Assert(res.Data(), HasLen, int(size))
		c.Assert(res.Data(), DeepEquals, []byte(testBlob[offset:offset+size]))
	}
}

// TestReadWrite writes a key with a data, then read it into buffer
func (s *SessionSuite) TestWriteReadInto(c *C) {
	var (
		testBlob                 = `MY_TEST_BLOB_WITH_DUMMY_DATA`
		testKey                  = fmt.Sprintf("testkey-%d", time.Now().Unix())
		testNamespace            = fmt.Sprintf("testnamespace-%d", time.Now().Unix())
		testBlobReader io.Reader = strings.NewReader(testBlob)
	)

	s.session.SetGroups(s.groups)
	s.session.SetNamespace(testNamespace)

	for res := range s.session.WriteData(testKey, testBlobReader, 0, 0) {
		c.Assert(res.Error(), IsNil)
	}

	// No overflow must be there
	// TODO: addd Assert
	offset := uint64(2)
	p := make([]byte, 2)
	key, _ := NewKey(testKey)
	defer key.Free()

	for res := range s.session.ReadInto(key, offset, p) {
		// No read error
		c.Assert(res.Error(), IsNil)
		// Read exactly size
		c.Assert(res.Data(), HasLen, len(p))
		c.Assert(p, DeepEquals, []byte(testBlob[offset:offset+uint64(len(p))]))
	}
}

// TestLookupWriteLookup lookups a random key, writes it, then looks it up again
func (s *SessionSuite) TestLookupWriteLookup(c *C) {
	// TODO: copy-paste
	var (
		testBlob                 = `MY_TEST_BLOB_WITH_DUMMY_DATA`
		testKey                  = fmt.Sprintf("testkey-%d", time.Now().Unix())
		testNamespace            = fmt.Sprintf("testnamespace-%d", time.Now().Unix())
		testBlobReader io.Reader = strings.NewReader(testBlob)
	)

	s.session.SetGroups(s.groups)
	s.session.SetNamespace(testNamespace)

	key, _ := NewKey(testKey)
	defer key.Free()

	for res := range s.session.Lookup(key) {
		c.Assert(res.Error(), ErrorMatches, ".*Failed to process LOOKUP command: No such file or directory: -2.*$")
	}

	for res := range s.session.ParallelLookup(testKey) {
		c.Assert(res.Error(), ErrorMatches, ".*Failed to process LOOKUP command: No such file or directory: -2.*$")
	}

	for res := range s.session.WriteData(testKey, testBlobReader, 0, 0) {
		c.Assert(res.Error(), IsNil)
	}

	for res := range s.session.Lookup(key) {
		// No error
		c.Assert(res.Error(), IsNil)
		// Address must not be empty
		c.Assert(res.Addr().String(), Not(HasLen), 0)
		c.Assert(res.Info().Size, DeepEquals, uint64(len(testBlob)))
	}

	for res := range s.session.ParallelLookup(testKey) {
		// No error
		c.Assert(res.Error(), IsNil)
		c.Assert(res.Addr().String(), Not(HasLen), 0)
		c.Assert(res.Info().Size, DeepEquals, uint64(len(testBlob)))
	}
}

func (s *SessionSuite) TestWriteRemove(c *C) {
	// TODO: copy-paste
	var (
		testBlob                 = `MY_TEST_BLOB_WITH_DUMMY_DATA`
		testKey                  = fmt.Sprintf("testkey-%d", time.Now().Unix())
		testNamespace            = fmt.Sprintf("testnamespace-%d", time.Now().Unix())
		testBlobReader io.Reader = strings.NewReader(testBlob)
	)

	s.session.SetGroups(s.groups)
	s.session.SetNamespace(testNamespace)

	key, _ := NewKey(testKey)
	defer key.Free()

	for res := range s.session.WriteData(testKey, testBlobReader, 0, 0) {
		c.Assert(res.Error(), IsNil)
	}

	for res := range s.session.Remove(testKey) {
		c.Assert(res.Error(), IsNil)
	}

	for res := range s.session.Lookup(key) {
		// Check that the key was actually removed
		c.Assert(res.Error(), ErrorMatches, ".*Failed to process LOOKUP command: No such file or directory: -2.*$")
	}
}
