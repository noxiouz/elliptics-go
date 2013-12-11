package elliptics

import "C"

import (
	"unsafe"
)

var goCallback = GoCallback

//export GoCallback
func GoCallback(result unsafe.Pointer, context unsafe.Pointer) {
	callback := *(*func(unsafe.Pointer))(context)
	callback(result)
}
