package elliptics

import "C"

import (
	"fmt"
	"unsafe"
)

var goCallback = GoCallback

//export GoCallback
func GoCallback(result unsafe.Pointer, context unsafe.Pointer) {
	fmt.Println("Context pointer", context)
	callback := *(*func(unsafe.Pointer))(context)
	callback(result)
}
