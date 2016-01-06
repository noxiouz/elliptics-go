package elliptics

import (
	"time"

	. "gopkg.in/check.v1"
)

func init() {
	Suite(&NodeSuite{})
	Suite(&LoggerSuite{})
}

type LoggerSuite struct{}

func (s *LoggerSuite) TestLogLevelParsing(c *C) {
	_, err := NewNode("/dev/stderr", "invalidloglevel")
	c.Assert(err, ErrorMatches, "could not create node, please check stderr output")
}

type NodeSuite struct {
	node *Node
}

func (s *NodeSuite) SetUpTest(c *C) {
	node, err := NewNode("/dev/stderr", "info")
	c.Assert(err, IsNil)
	s.node = node
}

func (s *NodeSuite) TearDownTest(c *C) {
	// NOTE: to avoid abortion
	/*
		could not create new node: exception: Unknown log level: invalidloglevel
		2016-01-06 12:09:24.239103 0000000000000000/7656/7656 INFO: Elliptics starts, flags: 0x0 [], attrs: []
		2016-01-06 12:09:24.239879 0000000000000000/7656/7656 INFO: Grew BLOCKING pool by: 0 -> 8 IO threads, attrs: []
		2016-01-06 12:09:24.245673 0000000000000000/7656/7656 INFO: Grew NONBLOCKING pool by: 0 -> 4 IO threads, attrs: []
		terminate called after throwing an instance of 'boost::exception_detail::clone_impl<boost::exception_detail::error_info_injector<boost::lock_error> >'
		  what():  boost::lock_error
		signal: aborted
	*/
	time.Sleep(1 * time.Second)
	if s.node != nil {
		s.node.Free()
	}
}

func (s *NodeSuite) TestNodeAddRemote(c *C) {
	const malformedAddress = "blabla:1025:22"

	err := s.node.AddRemote(malformedAddress)
	c.Assert(err, ErrorMatches, "no such device or address")

	err = s.node.AddRemotes([]string{malformedAddress, malformedAddress})
	c.Assert(err, ErrorMatches, "invalid argument")

}

func (s *NodeSuite) TestAddEmptyRemotesList(c *C) {
	err := s.node.AddRemotes([]string{})
	c.Assert(err, ErrorMatches, "list of remotes is empty")
}
