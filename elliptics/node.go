package elliptics

// #cgo LDFLAGS: -lell -lelliptics_cpp -L .
// #include "node.h"
// #include <stdlib.h>
import "C"

import (
	"unsafe"
)

type Node struct {
	logger *Logger
	node   unsafe.Pointer
}

func NewNode(log *Logger) (node *Node, err error) {
	cnode, err := C.new_node(log.logger)
	if err != nil {
		return
	}
	node = &Node{log, cnode}
	return
}
