package elliptics

import (
	"io/ioutil"
	"os"
	"time"

	. "gopkg.in/check.v1"
)

func init() {
	Suite(&NodeSuite{})
	Suite(&LoggerSuite{})
}

type LoggerSuite struct{}

func (s *LoggerSuite) TestLogLevelParsing(c *C) {
	_, err := NewNode(os.DevNull, "invalidloglevel")
	c.Assert(err, ErrorMatches, "could not create node, please check stderr output")
}

type NodeSuite struct {
	logfile *os.File
	node    *Node
}

func (s *NodeSuite) SetUpTest(c *C) {
	file, err := ioutil.TempFile("", "elliptics-node-test-log.log")
	c.Assert(err, IsNil)
	node, err := NewNode(file.Name(), "info")
	c.Assert(err, IsNil)
	s.node = node
	s.logfile = file
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

	if s.logfile != nil {
		s.logfile.Close()
		if !c.Failed() {
			os.RemoveAll(s.logfile.Name())
		} else {
			c.Logf("you can find node logfile: %s", s.logfile.Name())
		}
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

func (s *NodeSuite) TestSetTimeouts(c *C) {
	s.node.SetTimeouts(100, 200)
}
