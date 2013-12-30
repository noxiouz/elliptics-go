package elliptics

//#include "gologger.h"
import "C"
import "unsafe"

//export GoLog
func GoLog(fptr, priv unsafe.Pointer, level int, msg *C.char) {
	f := *(*func(unsafe.Pointer, int, *C.char))(fptr)
	f(priv, level, msg)
}

func NewNodeLog(logf, priv unsafe.Pointer, level int) (node *Node, err error) {
	cnode, err := C.gologger_create(logf, priv, C.int(level))
	if err != nil {
		return
	}

	node = &Node{node: cnode}
	return
}
