package elliptics

//#include "gologger.h"
import "C"
import "unsafe"

//export GoLog
func GoLog(fptr, priv unsafe.Pointer, level int, msg *C.char) {
	f := *(*func(unsafe.Pointer, int, *C.char))(fptr)
	f(priv, level, msg)
}

func NewNode1(logger *Logger, level string) (node *Node, err error) {
	//cnode, err := C.gologger_create(C.int(level))
	if err != nil {
		return
	}

	node = &Node{node: nil}
	return
}
