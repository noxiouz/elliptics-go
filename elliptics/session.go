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

const VOLUME = 10

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
	Lookup() []LookupResult
	Error() error
}

type writeDataResult struct {
	lookup []LookupResult
	err    error
}

func (w *writeDataResult) Error() error {
	return w.err
}

func (w *writeDataResult) Lookup() []LookupResult {
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

//Set groups to the session
func (s *Session) SetGroups(groups []int32) {
	C.session_set_groups(s.session, (*C.int32_t)(&groups[0]), C.int(len(groups)))
}

//Set namespace for elliptics session.
//Default namespace is empty string.
func (s *Session) SetNamespace(namespace string) {
	cnamespace := C.CString(namespace)
	defer C.free(unsafe.Pointer(cnamespace))
	C.session_set_namespace(s.session, cnamespace, C.int(len(namespace)))
}

func (s *Session) ReadKey(key *Key) (responseCh chan IReadDataResult) {
	//Context is closure, which contains channel to answer in.
	//It will pass as the last argument to exported go_*_callback
	//through C++ callback after operation finish comes.
	//go_read_callback casts context to properly go func,
	//and calls with []ReadResult
	responseCh = make(chan IReadDataResult, 1)
	context := func(results []ReadResult, err int) {
		if err != 0 {
			responseCh <- &readDataResult{
				err: fmt.Errorf("%v", err),
				res: nil}
		} else {
			responseCh <- &readDataResult{err: nil, res: results}
		}
	}
	C.session_read_data(s.session, unsafe.Pointer(&context), key.key)
	return
}

func (s *Session) ReadData(key string) (responseCh chan IReadDataResult) {
	ekey, err := NewKey(key)
	if err != nil {
		errCh := make(chan IReadDataResult, 1)
		errCh <- &readDataResult{err, nil}
		return errCh
	}
	defer ekey.Free()
	return s.ReadKey(ekey)
}

func (s *Session) WriteData(key string, blob string) (responseCh chan IWriteDataResult) {
	ekey, err := NewKey(key)
	if err != nil {
		return
	}
	defer ekey.Free()
	return s.WriteKey(ekey, blob)
}

func (s *Session) WriteKey(key *Key, blob string) (responseCh chan IWriteDataResult) {
	//Similary to ReadKey
	responseCh = make(chan IWriteDataResult, VOLUME)
	raw_data := C.CString(blob) // Mustn't call free. Elliptics does it.
	context := func(result []LookupResult, err int) {
		if err != 0 {
			responseCh <- &writeDataResult{
				err:    fmt.Errorf("%v", err),
				lookup: nil}
		} else {
			responseCh <- &writeDataResult{
				err:    nil,
				lookup: result}
		}
	}
	C.session_write_data(s.session, unsafe.Pointer(&context), key.key, raw_data, C.size_t(len(blob)))
	return
}

func (s *Session) Remove(key string) (responseCh chan IRemoveResult) {
	ekey, err := NewKey(key)
	if err != nil {
		return
	}
	defer ekey.Free()
	return s.RemoveKey(ekey)
}

func (s *Session) RemoveKey(key *Key) (responseCh chan IRemoveResult) {
	responseCh = make(chan IRemoveResult, VOLUME)
	context := func(err int) {
		responseCh <- &removeResult{err: fmt.Errorf("%v", err)}
	}

	C.session_remove(s.session, unsafe.Pointer(&context), key.key)
	return
}

//Find interface
type Finder interface {
	Error() error
	Data() []IndexEntry
}

type FindResult struct {
	id   C.struct_dnet_raw_id
	data []IndexEntry
	err  error
}

type IndexEntry struct {
	Data string
}

func (f *FindResult) Data() []IndexEntry {
	return f.data
}

func (f *FindResult) Error() error {
	return f.err
}

func (s *Session) FindAllIndexes(indexes []string) <-chan Finder {
	responseCh := make(chan Finder, VOLUME)
	onResult, onFinish, cindexes := s.findIndexes(indexes, responseCh)
	C.session_find_all_indexes(s.session, onResult, onFinish,
		(**C.char)(&cindexes[0]), C.size_t(len(indexes)))
	return responseCh
}

func (s *Session) FindAnyIndexes(indexes []string) <-chan Finder {
	responseCh := make(chan Finder, VOLUME)
	onResult, onFinish, cindexes := s.findIndexes(indexes, responseCh)
	C.session_find_any_indexes(s.session, onResult, onFinish,
		(**C.char)(&cindexes[0]), C.size_t(len(indexes)))
	return responseCh
}

func (s *Session) findIndexes(indexes []string, responseCh chan Finder) (onResult, onFinish unsafe.Pointer, cindexes []*C.char) {
	for _, index := range indexes {
		cindex := C.CString(index)
		cindexes = append(cindexes, cindex)
	}

	_result := func(result *FindResult) {
		responseCh <- result
	}
	onResult = unsafe.Pointer(&_result)

	_finish := func(err int) {
		if err != 0 {
			responseCh <- &FindResult{err: fmt.Errorf("%d", err)}
		}
		close(responseCh)
	}
	onFinish = unsafe.Pointer(&_finish)
	return
}
