package elliptics

/*
#include "session.h"
*/
import "C"

import (
	"reflect"
	"unsafe"
)

type WriteResult struct {
	info C.struct_dnet_file_info //dnet_file_info
	addr C.struct_dnet_addr
	path string //file_path
}

//export go_write_callback
func go_write_callback(result *C.struct_go_write_result, size int, err int, context unsafe.Pointer) {
	callback := *(*func([]WriteResult, int))(context)
	if err != 0 {
		callback(nil, err)
	} else {
		var tmp []C.struct_go_write_result
		sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&tmp)))
		sliceHeader.Cap = size
		sliceHeader.Len = size
		sliceHeader.Data = uintptr(unsafe.Pointer(result))

		var Results []WriteResult
		for _, item := range tmp {
			Results = append(Results, WriteResult{
				info: *item.info,
				addr: *item.addr,
				path: C.GoString(item.path)})
		}
		// All data from cpp has been copied here.
		defer callback(Results, err)
	}

}

type ReadResult struct {
	Data string
}

//export go_read_callback
func go_read_callback(result *C.struct_go_read_result, size int, err int, context unsafe.Pointer) {
	callback := *(*func([]ReadResult, int))(context)
	if err != 0 {
		callback(nil, err)
	} else {
		var tmp []C.struct_go_read_result
		sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&tmp)))
		sliceHeader.Cap = size
		sliceHeader.Len = size
		sliceHeader.Data = uintptr(unsafe.Pointer(result))

		var Results []ReadResult
		for _, item := range tmp {
			Results = append(Results, ReadResult{C.GoString(item.file)})
		}
		// All data from cpp has been copied here.
		callback(Results, err)
	}
}
