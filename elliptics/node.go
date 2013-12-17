package elliptics

// #include "node.h"
// #include <stdlib.h>
import "C"

import (
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

func (node *Node) AddRemote(addr string) (err error) {
	caddr := C.CString(addr)
	defer C.free(unsafe.Pointer(caddr))

	_, c_err := C.node_add_remote_one(node.node, caddr)
	if c_err != nil {
		if err, ok := c_err.(syscall.Errno); ok && isError(err) {
			return err
		} else if !ok {
			return nil
		}
	}
	return
}
