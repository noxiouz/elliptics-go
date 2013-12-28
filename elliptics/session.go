/*
* 2013+ Copyright (c) Anton Tyurin <noxiouz@yandex.ru>
* All rights reserved.
*
* This program is free software; you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation; either version 2 of the License, or
* (at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU General Public License for more details.
 */

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

const defaultVOLUME = 10

const (
	indexesSet = iota
	indexesUpdate
)

//Session
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

/*
	Read
*/

type ReadResult interface {
	Data() string
	Error() error
}

type readResult struct {
	ioAttr C.struct_dnet_io_attr
	data   string
	err    error
}

func (r *readResult) Data() string {
	return r.data
}

func (r *readResult) Error() error {
	return r.err
}

func (s *Session) ReadKey(key *Key) <-chan ReadResult {
	//Context is closure, which contains channel to answer in.
	//It will pass as the last argument to exported go_*_callback
	//through C++ callback after operation finish comes.
	//go_read_callback casts context to properly go func,
	//and calls with []ReadResult
	responseCh := make(chan ReadResult, defaultVOLUME)
	onResult := func(result readResult) {
		responseCh <- &result
	}

	onFinish := func(err int) {
		if err != 0 {
			responseCh <- &readResult{err: fmt.Errorf("%d", err)}
		}
		close(responseCh)
	}
	C.session_read_data(s.session,
		unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
		key.key)
	return responseCh
}

func (s *Session) ReadData(key string) <-chan ReadResult {
	ekey, err := NewKey(key)
	if err != nil {
		errCh := make(chan ReadResult, 1)
		errCh <- &readResult{err: err}
		close(errCh)
		return errCh
	}
	defer ekey.Free()
	return s.ReadKey(ekey)
}

/*
	Write and Lookup
*/

type Lookuper interface {
	Path() string
	Addr() C.struct_dnet_addr
	Info() C.struct_dnet_file_info
	Error() error
}

type lookupResult struct {
	info C.struct_dnet_file_info //dnet_file_info
	addr C.struct_dnet_addr
	path string //file_path
	err  error
}

func (l *lookupResult) Path() string {
	return l.path
}

func (l *lookupResult) Addr() C.struct_dnet_addr {
	return l.addr
}

func (l *lookupResult) Info() C.struct_dnet_file_info {
	return l.info
}

func (l *lookupResult) Error() error {
	return l.err
}

func (s *Session) WriteData(key string, blob string) <-chan Lookuper {
	ekey, err := NewKey(key)
	if err != nil {
		responseCh := make(chan Lookuper, defaultVOLUME)
		responseCh <- &lookupResult{err: err}
		close(responseCh)
		return responseCh
	}
	defer ekey.Free()
	return s.WriteKey(ekey, blob)
}

func (s *Session) WriteKey(key *Key, blob string) <-chan Lookuper {
	responseCh := make(chan Lookuper, defaultVOLUME)
	raw_data := C.CString(blob) // Mustn't call free. Elliptics does it.

	onResult := func(lookup *lookupResult) {
		responseCh <- lookup
	}

	onFinish := func(err int) {
		if err != 0 {
			responseCh <- &lookupResult{err: fmt.Errorf("%d", err)}
		}
		close(responseCh)
	}

	C.session_write_data(s.session,
		unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
		key.key, raw_data, C.size_t(len(blob)))
	return responseCh
}

func (s *Session) Lookup(key *Key) <-chan Lookuper {
	responseCh := make(chan Lookuper, defaultVOLUME)

	onResult := func(lookup *lookupResult) {
		responseCh <- lookup
	}

	onFinish := func(err int) {
		if err != 0 {
			responseCh <- &lookupResult{err: fmt.Errorf("%d", err)}
		}
		close(responseCh)
	}

	C.session_lookup(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish), key.key)
	return responseCh
}

/*
	Remove
*/

type Remover interface {
	Error() error
}

type removeResult struct {
	err error
}

func (r *removeResult) Error() error {
	return r.err
}

func (s *Session) Remove(key string) <-chan Remover {
	ekey, err := NewKey(key)
	if err != nil {
		responseCh := make(chan Remover, defaultVOLUME)
		responseCh <- &removeResult{err: err}
		close(responseCh)
		return responseCh
	}
	defer ekey.Free()
	return s.RemoveKey(ekey)
}

func (s *Session) RemoveKey(key *Key) <-chan Remover {
	responseCh := make(chan Remover, defaultVOLUME)
	onResult := func() {
		//It's never called.
	}
	onFinish := func(err int) {
		if err != 0 {
			responseCh <- &removeResult{err: fmt.Errorf("%v", err)}
		}
		close(responseCh)
	}

	C.session_remove(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish), key.key)
	return responseCh
}

/*
	Find
*/

type Finder interface {
	Error() error
	Data() []IndexEntry
}

type findResult struct {
	id   C.struct_dnet_raw_id
	data []IndexEntry
	err  error
}

type IndexEntry struct {
	Data string
	err  error
}

func (f *findResult) Data() []IndexEntry {
	return f.data
}

func (f *findResult) Error() error {
	return f.err
}

func (s *Session) FindAllIndexes(indexes []string) <-chan Finder {
	responseCh := make(chan Finder, defaultVOLUME)
	onResult, onFinish, cindexes := s.findIndexes(indexes, responseCh)
	C.session_find_all_indexes(s.session, onResult, onFinish,
		(**C.char)(&cindexes[0]), C.size_t(len(indexes)))
	//Free cindexes
	for _, item := range cindexes {
		C.free(unsafe.Pointer(item))
	}
	return responseCh
}

func (s *Session) FindAnyIndexes(indexes []string) <-chan Finder {
	responseCh := make(chan Finder, defaultVOLUME)
	onResult, onFinish, cindexes := s.findIndexes(indexes, responseCh)
	C.session_find_any_indexes(s.session, onResult, onFinish,
		(**C.char)(&cindexes[0]), C.size_t(len(indexes)))
	//Free cindexes
	for _, item := range cindexes {
		C.free(unsafe.Pointer(item))
	}
	return responseCh
}

func (s *Session) findIndexes(indexes []string, responseCh chan Finder) (onResult, onFinish unsafe.Pointer, cindexes []*C.char) {
	for _, index := range indexes {
		cindex := C.CString(index)
		cindexes = append(cindexes, cindex)
	}

	_result := func(result *findResult) {
		responseCh <- result
	}
	onResult = unsafe.Pointer(&_result)

	_finish := func(err int) {
		if err != 0 {
			responseCh <- &findResult{err: fmt.Errorf("%d", err)}
		}
		close(responseCh)
	}
	onFinish = unsafe.Pointer(&_finish)
	return
}

/*
	Indexes
*/

type Indexer interface {
	Error() error
}

type indexResult struct {
	err error
}

func (i *indexResult) Error() error {
	return i.err
}

func (s *Session) setOrUpdateIndexes(operation int, key string, indexes map[string]string) <-chan Indexer {
	ekey, err := NewKey(key)
	if err != nil {
		panic(err)
	}
	defer ekey.Free()
	responseCh := make(chan Indexer, defaultVOLUME)

	var cindexes []*C.char
	var cdatas []C.struct_go_data_pointer

	for index, data := range indexes {
		cindex := C.CString(index) // free this
		defer C.free(unsafe.Pointer(cindex))
		cindexes = append(cindexes, cindex)

		cdata := C.new_data_pointer(
			C.CString(data), // freed by ellipics::data_pointer in std::vector ???
			C.int(len(data)),
		)
		cdatas = append(cdatas, cdata)
	}

	onResult := func() {
		//It's never called. For the future.
	}

	onFinish := func(err int) {
		if err != 0 {
			responseCh <- &indexResult{err: fmt.Errorf("%v", err)}
		}
		close(responseCh)
	}
	// TODO: Reimplement this with pointer on functions
	switch operation {
	case indexesSet:
		C.session_set_indexes(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
			ekey.key,
			(**C.char)(&cindexes[0]),
			(*C.struct_go_data_pointer)(&cdatas[0]),
			C.size_t(len(cindexes)))

	case indexesUpdate:
		C.session_update_indexes(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
			ekey.key,
			(**C.char)(&cindexes[0]),
			(*C.struct_go_data_pointer)(&cdatas[0]),
			C.size_t(len(cindexes)))
	}
	return responseCh
}

func (s *Session) SetIndexes(key string, indexes map[string]string) <-chan Indexer {
	return s.setOrUpdateIndexes(indexesSet, key, indexes)
}

func (s *Session) UpdateIndexes(key string, indexes map[string]string) <-chan Indexer {
	return s.setOrUpdateIndexes(indexesUpdate, key, indexes)
}

func (s *Session) ListIndexes(key string) <-chan IndexEntry {
	responseCh := make(chan IndexEntry, defaultVOLUME)
	ekey, err := NewKey(key)
	if err != nil {
		panic(err)
	}
	defer ekey.Free()

	onResult := func(indexentry *IndexEntry) {
		responseCh <- *indexentry
	}

	onFinish := func(err int) {
		if err != 0 {
			responseCh <- IndexEntry{err: fmt.Errorf("%v", err)}
		}
		close(responseCh)
	}

	C.session_list_indexes(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish), ekey.key)
	return responseCh
}

func (s *Session) RemoveIndexes(key string, indexes []string) <-chan Indexer {
	ekey, err := NewKey(key)
	if err != nil {
		panic(err)
	}
	defer ekey.Free()
	responseCh := make(chan Indexer, defaultVOLUME)

	var cindexes []*C.char
	for _, index := range indexes {
		cindex := C.CString(index) // free this
		defer C.free(unsafe.Pointer(cindex))
		cindexes = append(cindexes, cindex)
	}

	onResult := func() {
		//It's never called. For the future.
	}

	onFinish := func(err int) {
		if err != 0 {
			responseCh <- &indexResult{err: fmt.Errorf("%v", err)}
		}
		close(responseCh)
	}

	C.session_remove_indexes(s.session,
		unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
		ekey.key, (**C.char)(&cindexes[0]), C.size_t(len(cindexes)))
	return responseCh
}
