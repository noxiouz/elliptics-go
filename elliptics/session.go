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

var f []interface{}

// Result of read operation
type IReadDataResult interface {
	Error() error
	Data() []ReadResult
}

type readDataResult struct {
	err error
	res []ReadResult
}

func (r *readDataResult) Error() error {
	return r.err
}

func (r *readDataResult) Data() []ReadResult {
	return r.res
}

//Result of write operation
type IWriteDataResult interface {
	Lookup() []WriteResult
	Error() error
}

type writeDataResult struct {
	lookup []WriteResult
	err    error
}

func (w *writeDataResult) Error() error {
	return w.err
}

func (w *writeDataResult) Lookup() []WriteResult {
	return w.lookup
}

//Result of remove
type IRemoveResult interface {
	Error() error
}

type removeResult struct {
	err error
}

func (r *removeResult) Error() error {
	return r.err
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
	context := func(results []ReadResult, err int) {
		if err != 0 {
			a <- &readDataResult{
				err: fmt.Errorf("%v", err),
				res: nil}
		} else {
			a <- &readDataResult{err: nil, res: results}
		}
	}
	C.session_read_data(s.session, unsafe.Pointer(&context), key.key)
	return
}

func (s *Session) ReadData(key string) (a chan IReadDataResult) {
	ekey, err := NewKey(key)
	if err != nil {
		errCh := make(chan IReadDataResult, 1)
		errCh <- &readDataResult{err, nil}
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
	context := func(result []WriteResult, err int) {
		if err != 0 {
			a <- &writeDataResult{
				err:    fmt.Errorf("%v", err),
				lookup: nil}
		} else {
			a <- &writeDataResult{
				err:    nil,
				lookup: result}
		}
	}

	raw_data := C.CString(blob)
	defer C.free(unsafe.Pointer(raw_data))
	C.session_write_data(s.session, unsafe.Pointer(&context), key.key, raw_data, C.size_t(len(blob)))
	return
}

func (s *Session) Remove(key string) (a chan IRemoveResult) {
	ekey, err := NewKey(key)
	if err != nil {
		return
	}
	defer ekey.Free()
	return s.RemoveKey(ekey)
}

func (s *Session) RemoveKey(key *Key) (a chan IRemoveResult) {
	a = make(chan IRemoveResult, 1)
	context := func(err int) {
		a <- &removeResult{err: fmt.Errorf("%v", err)}
	}

	C.session_remove(s.session, unsafe.Pointer(&context), key.key)
	return
}
