package elliptics

/*
#include "session.h"
*/
import "C"

import (
	"reflect"
	"unsafe"
)

type LookupResult struct {
	info C.struct_dnet_file_info //dnet_file_info
	addr C.struct_dnet_addr
	path string //file_path
}

//export go_lookup_callback
func go_lookup_callback(result *C.struct_go_write_result, size int, err int, context unsafe.Pointer) {
	callback := *(*func([]LookupResult, int))(context)
	if err != 0 {
		callback(nil, err)
	} else {
		var tmp []C.struct_go_write_result
		sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&tmp)))
		sliceHeader.Cap = size
		sliceHeader.Len = size
		sliceHeader.Data = uintptr(unsafe.Pointer(result))

		var Results []LookupResult
		for _, item := range tmp {
			Results = append(Results, LookupResult{
				info: *item.info,
				addr: *item.addr,
				path: C.GoString(item.path)})
		}
		// All data from cpp has been copied here.
		defer callback(Results, err)
	}
}

//export go_final_callback
func go_final_callback(err int, context unsafe.Pointer) {
	callback := *(*func(int))(context)
	callback(err)
}

//export go_read_callback
func go_read_callback(item *C.struct_go_read_result, context unsafe.Pointer) {
	callback := *(*func(readResult))(context)

	Result := readResult{
		data:   C.GoStringN(item.file, C.int(item.size)),
		ioAttr: *item.io_attribute,
		err:    nil,
	}
	// All data from C++ has been copied here.
	callback(Result)
}

//export go_remove_callback
func go_remove_callback(err int, context unsafe.Pointer) {
	callback := *(*func(int))(context)
	callback(err)
}

//export go_find_callback
func go_find_callback(result *C.struct_go_find_result, context unsafe.Pointer) {
	callback := *(*func(*FindResult))(context)
	var indexEntries []C.struct_c_index_entry
	size := int(result.entries_count)
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&indexEntries)))
	sliceHeader.Cap = size
	sliceHeader.Len = size
	sliceHeader.Data = uintptr(unsafe.Pointer(result.entries))
	var IndexDatas []IndexEntry
	for _, item := range indexEntries {
		IndexDatas = append(IndexDatas, IndexEntry{
			Data: C.GoStringN(item.data, C.int(item.size)),
		})
	}
	callback(&FindResult{
		id:   *result.id,
		data: IndexDatas,
		err:  nil,
	})
}
