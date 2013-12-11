package elliptics

import (
	"fmt"
	"unsafe"
)

/*
#include "session.h"
#include <stdio.h>
*/
import "C"

var _ = fmt.Scanf

// Result of read operation
type IReadDataResult interface {
	Error() error
	Data() string
}

type readDataResult struct {
	err error
	res string
}

func (r *readDataResult) Error() error {
	return r.err
}

func (r *readDataResult) Data() string {
	return r.res
}

//Result of write operation
type IWriteDataResult interface {
	Error() error
}

type writeDataResult struct {
	err error
}

func (w *writeDataResult) Error() error {
	return w.err
}

// Session context
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

//Set elliptics groups for the session
func (s *Session) SetGroups(groups []int32) {
	C.session_set_groups(s.session, (*C.int32_t)(&groups[0]), C.int(len(groups)))
}

func (s *Session) ReadKey(key *Key) (a chan IReadDataResult) {
	a = make(chan IReadDataResult, 1)
	context := func(result unsafe.Pointer) {
		resa := (*C.struct_GoRes)(result)
		defer C.free(result)
		err := C.int(resa.errcode)
		if err != 0 {
			a <- &readDataResult{err: fmt.Errorf("%s", C.GoString((*C.char)(resa.result))),
				res: ""}
		} else {
			res := C.GoString((*C.char)(resa.result))
			a <- &readDataResult{err: nil, res: res}
		}
	}

	C.session_read_data(s.session, unsafe.Pointer(&context), key.key)
	return
}

func (s *Session) ReadData(key string) (a chan IReadDataResult) {
	ekey, err := NewKey(key)
	if err != nil {
		errCh := make(chan IReadDataResult, 1)
		errCh <- &readDataResult{err, ""}
		return errCh
	}
	defer ekey.Free()
	return s.ReadKey(ekey)
}

func (s *Session) WriteData(key string, blob string) (a chan IWriteDataResult) {
	ekey, err := NewKey(key)
	if err != nil {
		return
	}
	defer ekey.Free()
	return s.WriteKey(ekey, blob)
}

func (s *Session) WriteKey(key *Key, blob string) (a chan IWriteDataResult) {
	a = make(chan IWriteDataResult, 1)
	context := func(result unsafe.Pointer) {
		resa := (*C.struct_GoRes)(result)
		//defer C.free(result)
		err := C.int(resa.errcode)
		if err != 0 {
			errmsg := C.GoString((*C.char)(resa.result))
			a <- &readDataResult{err: fmt.Errorf("%s", errmsg),
				res: ""}
		} else {
			a <- &readDataResult{err: nil, res: ""}
		}
	}

	raw_data := C.CString(blob)
	defer C.free(unsafe.Pointer(raw_data))
	C.session_write_data(s.session, unsafe.Pointer(&context), key.key, raw_data, C.size_t(len(blob)))
	return
}
