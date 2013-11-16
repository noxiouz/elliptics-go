package elliptics

// #cgo LDFLAGS: -lell -lelliptics_cpp -L .
// #include "node.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"net"
	"strconv"
	"syscall"
	"unsafe"
)

type Node struct {
	logger *Logger
	node   unsafe.Pointer
}

func isError(errno syscall.Errno) bool {
	return errno != syscall.EINPROGRESS &&
		errno != syscall.EAGAIN &&
		errno != syscall.EALREADY &&
		errno != syscall.EISCONN
}

func NewNode(log *Logger) (node *Node, err error) {
	cnode, err := C.new_node(log.logger)
	if err != nil {
		return
	}
	node = &Node{log, cnode}
	return
}

func (node *Node) Free() {
	C.delete_node(node.node)
}

func (node *Node) SetTimeouts(waitTimeout int, checkTimeout int) {
	C.node_set_timeouts(node.node, C.int(waitTimeout), C.int(checkTimeout))
}

func (node *Node) AddRemote(args ...interface{}) (err error) {
	var addr string
	var sport string
	var port int
	var family int = syscall.AF_INET
	var ok bool = false

	switch count := len(args); count {
	case 1:
	case 2:
		family, ok = args[1].(int)
		if !ok {
			return errors.New("Wrong type of family argument. Use uint")
		}
	default:
		err = errors.New("Wrong arguments")
		return
	}

	if addr, ok = args[0].(string); !ok {
		return errors.New("Wrong type of endpoint argument")
	}

	addr, sport, err = net.SplitHostPort(addr)
	if err != nil {
		return
	}

	port, err = strconv.Atoi(sport)
	if err != nil {
		return
	}

	caddr := C.CString(addr)
	defer C.free(unsafe.Pointer(caddr))

	_, c_err := C.node_add_remote(node.node, caddr, C.int(port), C.int(family))
	if c_err != nil {
		if err, ok := c_err.(syscall.Errno); ok && isError(err) {
			return err
		} else if !ok {
			return nil
		}
	}
	return
}
