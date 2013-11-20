package elliptics

import (
	"fmt"
	"unsafe"
)

// #cgo LDFLAGS: -lell -lelliptics_cpp -L . 
// #include "session.h"
import "C"

//export Foo
func Foo() {
	fmt.Println("FOO")
}

type Session struct {
	session unsafe.Pointer
}

func NewSession(node *Node) (*Session, error) {
	session, err := C.new_elliptics_session(node.node)
	if err != nil {
		return nil, err
	}
	return &Session{session}, err
}
