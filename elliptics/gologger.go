package elliptics

import (
	"C"
	"log"
	"unsafe"
)

//export GoLog
func GoLog(priv unsafe.Pointer, msg *C.char) {
	var l *log.Logger = (*log.Logger)(priv)
	str := C.GoString(msg)
	l.Output(2, str)
}
