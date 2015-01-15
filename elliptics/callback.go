package elliptics

/*
#include "session.h"
*/
import "C"

import (
	"log"
	"reflect"
	"unsafe"
)

//export go_final_callback
func go_final_callback(cerr *C.struct_go_error, key uint64) {
	// callback := *(*func(error))(context)
	context, err := Pool.Get(uint64(key))
	if err != nil {
		panic("Unable to find session numbder")
	}
	callback := context.(func(error))

	log.Printf("go_final_callback: key: %d, context: %p, error_code: %d, error_message: %p\n",
		key, context, cerr.code, cerr.message)

	if cerr.code < 0 {
		err = &DnetError{
			Code:    int(cerr.code),
			Flags:   uint64(cerr.flags),
			Message: C.GoString(cerr.message),
		}

		callback(err)
	} else {
		callback(nil)
	}
}

//export go_lookup_error
func go_lookup_error(cmd *C.struct_dnet_cmd, addr *C.struct_dnet_addr, cerr *C.struct_go_error, key uint64) {
	// callback := *(*func(*lookupResult))(context)
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find session numbder")
	}
	callback := context.(func(*lookupResult))

	log.Printf("go_lookup_error: key: %d, context: %p, cmd: %p, addr: %p, error_code: %d, error_message: %p\n",
		key, context, cmd, addr, cerr.code, cerr.message)

	Result := lookupResult{
		cmd:  NewDnetCmd(cmd),
		addr: NewDnetAddr(addr),
		err: &DnetError{
			Code:    int(cerr.code),
			Flags:   uint64(cerr.flags),
			Message: C.GoString(cerr.message),
		},
	}
	callback(&Result)
}

//export go_lookup_callback
func go_lookup_callback(result *C.struct_go_lookup_result, key uint64) {
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find session numbder")
	}
	callback := context.(func(*lookupResult))

	log.Printf("go_lookup_callback: key: %d, context: %p, result.cmd: %p, result.addr: %p, result.info: %p, result.storage_addr: %p, result.path: %p\n",
		key, context, result.cmd, result.addr, result.info, result.storage_addr, result.path)

	Result := lookupResult{
		cmd:          NewDnetCmd(result.cmd),
		addr:         NewDnetAddr(result.addr),
		info:         NewDnetFileInfo(result.info),
		storage_addr: NewDnetAddr(result.storage_addr),
		path:         C.GoString(result.path),
		err:          nil,
	}
	callback(&Result)
}

//export go_remove_callback
func go_remove_callback(result *C.struct_go_remove_result, key uint64) {
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find session numbder")
	}
	callback := context.(func(*removeResult))

	log.Printf("go_remove_callback: key: %d, context: %p, result.cmd: %p\n",
		key, context, result.cmd)

	Result := removeResult{
		cmd: NewDnetCmd(result.cmd),
	}
	callback(&Result)
}

//export go_read_callback
func go_read_callback(result *C.struct_go_read_result, key uint64) {
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find session numbder")
	}
	callback := context.(func(*readResult))

	log.Printf("go_read_callback: key: %d, context: %p, result.cmd: %p, result.addr: %p, result.io: %p, result.file: %p, result.size: %d\n",
		key, context, result.cmd, result.addr, result.io_attribute, result.file, result.size)

	Result := &readResult{
		cmd:    NewDnetCmd(result.cmd),
		addr:   NewDnetAddr(result.addr),
		ioattr: NewDnetIOAttr(result.io_attribute),
		err:    nil,
	}

	if result.size > 0 && result.file != nil {
		Result.data = C.GoBytes(unsafe.Pointer(result.file), C.int(result.size))
	} else {
		Result.data = make([]byte, 0)
	}
	// All data from C++ has been copied here.
	callback(Result)
}

//export go_find_callback
func go_find_callback(result *C.struct_go_find_result, key uint64) {
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find session numbder")
	}
	callback := context.(func(*findResult))

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
	callback(&findResult{
		id:   *result.id,
		data: IndexDatas,
		err:  nil,
	})
}

//export go_index_entry_callback
func go_index_entry_callback(result *C.struct_c_index_entry, key uint64) {
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find session numbder")
	}
	callback := context.(func(*IndexEntry))

	callback(&IndexEntry{Data: C.GoStringN(result.data, C.int(result.size))})
}
