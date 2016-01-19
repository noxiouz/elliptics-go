package elliptics

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"
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
	ioserv  *DnetIOServ
}

func (s *SessionSuite) SetUpSuite(c *C) {
	// set groups
	s.groups = []uint32{1, 2, 3}

	// start ioserver
	ioserv, err := StartDnetIOServ(s.groups)
	if err != nil {
		c.Fatal(err)
	}
	s.ioserv = ioserv

	c.Logf("ioserv started [PID %d] on %s, groups %v",
		ioserv.cmd.Process.Pid, strings.Join(s.ioserv.Address(), ","), s.groups)
	// create node
	s.NodeSuite.SetUpTest(c)
	c.Logf("add remotes %s", strings.Join(s.ioserv.Address(), ","))
	s.node.AddRemotes(s.ioserv.Address())
}

func (s *SessionSuite) TearDownSuite(c *C) {
	s.NodeSuite.TearDownTest(c)

	if s.ioserv != nil {
		s.ioserv.Close()
	}
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

func (s *SessionSuite) TestTransform(c *C) {
	const key = "some_data"
	id := s.session.Transform(key)
	// NOTE: add more asserts
	c.Assert(id, Not(HasLen), 0)
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
	defer s.session.GetRoutes(dnetStat)
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
		sessionTraceID = TraceID(99999)
	)

	s.session.SetGroups(s.groups)
	s.session.SetTraceID(sessionTraceID)
	s.session.SetNamespace(testNamespace)

	key, _ := NewKey(testKey)
	defer key.Free()

	var backend int32

	for res := range s.session.WriteData(testKey, testBlobReader, 0, 0) {
		c.Assert(res.Error(), IsNil)

		dnetCmd := res.Cmd()
		c.Assert(dnetCmd, NotNil)
		c.Check(dnetCmd.Trace, DeepEquals, uint64(sessionTraceID))
		// TODO: dnetCmd.Flags -> uin64, should return IOflag
		c.Check(IOflag(dnetCmd.Flags), Equals, DNET_IO_FLAGS_NODATA)

		backend = res.Cmd().Backend
	}

	for res := range s.session.Remove(testKey) {
		c.Assert(res.Error(), IsNil)
		c.Check(res.Key(), Equals, testKey)

		dnetCmd := res.Cmd()
		c.Assert(dnetCmd, NotNil)

		c.Check(dnetCmd.Trace, DeepEquals, uint64(sessionTraceID))
		c.Check(dnetCmd.Flags, Equals, uint64(0))
		c.Check(dnetCmd.Backend, Equals, backend)
	}

	for res := range s.session.Lookup(key) {
		// Check that the key was actually removed
		c.Assert(res.Error(), ErrorMatches, ".*Failed to process LOOKUP command: No such file or directory: -2.*$")
	}
}

func (s *SessionSuite) TestWriteKeyZeroLengthError(c *C) {
	var testEmptyBlobReader io.Reader = strings.NewReader("")

	key, _ := NewKey("test-key")
	defer key.Free()

	for res := range s.session.WriteKey(key, testEmptyBlobReader, 0, 0) {
		c.Assert(res.Error(), NotNil)

		dnetErr, ok := res.Error().(*DnetError)
		c.Assert(ok, Equals, true)
		// NOTE: EINVAL
		c.Check(dnetErr.Code, Equals, -22)
		c.Check(dnetErr.Flags, Equals, uint64(0))
	}
}

func (s *SessionSuite) TestWriteKeyReaderError(c *C) {
	var testErrorReader = iotest.TimeoutReader(strings.NewReader("ABCD"))

	key, _ := NewKey("test-key")
	defer key.Free()

	for res := range s.session.WriteKey(key, testErrorReader, 0, 0) {
		c.Assert(res.Error(), Equals, iotest.ErrTimeout)
	}
}

func (s *SessionSuite) TestLookupBackend(c *C) {
	var group = s.groups[0]
	addr, backend, err := s.session.LookupBackend("test-key", group)
	c.Assert(err, IsNil)

	c.Check(backend, Equals, int32(group))

	// NOTE: assume IPv4 or hostname
	port := strings.Split(s.ioserv.Address()[0], ":")[1]

	// NOTE: test server is run on localhost and IPv4
	c.Check(addr.HostString(), Equals, "127.0.0.1")

	// 2 means IPv4
	c.Check(addr.Family, Equals, uint16(2))
	c.Check(addr.String(), Equals, fmt.Sprintf("127.0.0.1:%s:2", port))
}

func (s *SessionSuite) TestLookupBackendError(c *C) {
	const nonExistentGroup = 1000
	_, _, err := s.session.LookupBackend("test-key", nonExistentGroup)
	c.Assert(err, NotNil)
	c.Check(err, ErrorMatches, "elliptics error: -6: could not lookup backend: key '.*', group: .*: -6")

	dnetErr, ok := err.(*DnetError)
	c.Assert(ok, Equals, true)
	c.Check(dnetErr.Code, Equals, -6)
	c.Check(dnetErr.Flags, Equals, uint64(0))
}

// TestReadWrite writes a key with a data, then read it
func (s *SessionSuite) TestReader(c *C) {
	var (
		testBlob                 = `MY_TEST_BLOB_WITH_DUMMY_DATA`
		testKey                  = fmt.Sprintf("testkey-%d", time.Now().Unix())
		testNamespace            = fmt.Sprintf("testnamespace-%d", time.Now().Unix())
		testBlobReader io.Reader = strings.NewReader(testBlob)
	)

	s.session.SetGroups(s.groups)
	s.session.SetNamespace(testNamespace)

	// Store test data
	for res := range s.session.WriteData(testKey, testBlobReader, 0, 0) {
		c.Assert(res.Error(), IsNil)
	}

	var (
		offset = uint64(2)
		size = uint64(len(testBlob)) - offset
	)

	readSeeker, err := NewReadSeekerOffsetSize(s.session, testKey, offset, size)
	c.Assert(err, IsNil)
	defer readSeeker.Free()

	var buff = make([]byte, 2)
	n, err := readSeeker.Read(buff)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, len(buff))
	c.Assert(string(buff), DeepEquals, testBlob[int(offset):int(offset)+len(buff)])

	buf_size := 2
	buff = make([]byte, buf_size)

	readSeeker.Seek(int64(-buf_size) + 1, 2)
	n, err = readSeeker.Read(buff)
	c.Check(n, Equals, len(buff) - 1)
	c.Assert(err, IsNil)

	buff = buff[:n]
	c.Assert(string(buff), DeepEquals, testBlob[len(testBlob) - buf_size + 1:])

	readSeeker.Seek(0, 2)
	n, err = readSeeker.Read(buff)
	// we are at the very end of the file, reading should fail
	c.Check(err, Equals, io.EOF)
}
