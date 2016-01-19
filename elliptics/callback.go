package elliptics

/*
#include "session.h"
*/
import "C"

import (
	"unsafe"
)

//export go_final_callback
func go_final_callback(cerr *C.struct_go_error, key uint64) {
	context, err := Pool.Get(uint64(key))
	if err != nil {
		panic("Unable to find final callback")
	}
	callback := context.(func(error))

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
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find lookup callback")
	}
	callback := context.(func(*lookupResult))

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
		panic("Unable to find lookup callback")
	}
	callback := context.(func(*lookupResult))

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
		panic("Unable to find remove callback")
	}
	callback := context.(func(*removeResult))

	Result := removeResult{
		cmd: NewDnetCmd(result.cmd),
	}
	callback(&Result)
}

//export go_read_error
func go_read_error(cmd *C.struct_dnet_cmd, addr *C.struct_dnet_addr, cerr *C.struct_go_error, key uint64) {
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find read callback")
	}
	callback := context.(func(*readResult))

	Result := readResult{
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

//export go_read_callback
func go_read_callback(result *C.struct_go_read_result, key uint64, buffer_key uint64) {
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find read callback")
	}
	callback := context.(func(*readResult))

	Result := &readResult{
		cmd:    NewDnetCmd(result.cmd),
		addr:   NewDnetAddr(result.addr),
		ioattr: NewDnetIOAttr(result.io_attribute),
		err:    nil,
	}

	if buffer_key != 0 {
		buffer_context, err := Pool.Get(buffer_key)
		if err != nil {
			panic("Unable to find buffer key context")
		}
		buffer := buffer_context.([]byte)

		size := uint64(len(buffer))
		if uint64(result.size) < size {
			size = uint64(result.size)
		}

		C.memmove(unsafe.Pointer(&buffer[0]), unsafe.Pointer(result.file), C.size_t(size))
		Result.ioattr.Size = size
		Result.data = buffer[:size]
	} else {
		if result.size > 0 && result.file != nil {
			Result.data = C.GoBytes(unsafe.Pointer(result.file), C.int(result.size))
		} else {
			Result.data = make([]byte, 0)
		}

		Result.ioattr.Size = uint64(len(Result.data))
	}

	// All data from C++ has been copied here.
	callback(Result)
}
