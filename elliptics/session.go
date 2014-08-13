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

/*Session allows to perfom any operations with data and indexes.

Most of methods return channel. Channel will be closed after results end or
error occurs. In case of error last value received from channel returns non nil value
from Error method.

For example Remove:

	if rm, ok := <-session.Remove(KEY); !ok {
		//Remove normally doesn't return any value, so chanel was closed.
		log.Println("Remove successfully")
	} else {
		//We's received value from channel. It should be error message.
		log.Println("Error occured: ", rm.Error())
	}
*/
type Session struct {
	session unsafe.Pointer
}

//NewSession returns Session connected with given Node.
func NewSession(node *Node) (*Session, error) {
	session, err := C.new_elliptics_session(node.node)
	if err != nil {
		return nil, err
	}
	return &Session{session}, err
}

//SetGroups points groups Session should work with.
func (s *Session) SetGroups(groups []int32) {
	C.session_set_groups(s.session, (*C.int32_t)(&groups[0]), C.int(len(groups)))
}

//SetTimeout sets wait timeout in seconds (time to wait for operation to complete) for all subsequent session operations
func (s *Session) SetTimeout(timeout int) {
	C.session_set_timeout(s.session, C.int(timeout))
}

//SetCflags sets command flags (DNET_FLAGS_* in API documentation) like nolock
func (s *Session) SetCflags(cflags uint64) {
	C.session_set_cflags(s.session, C.uint64_t(cflags))
}

//SetIOflags sets IO flags (DNET_IO_FLAGS_* in API documentation), i.e. flags for IO operations like read/write/delete
func (s *Session) SetIOflags(ioflags uint32) {
	C.session_set_ioflags(s.session, C.uint32_t(ioflags))
}

func (s *Session) SetTraceID(trace uint64) {
	C.session_set_trace_id(s.session, C.uint64_t(trace))
}

/*SetNamespace sets the namespace for the Session.Default namespace is empty string.

This feature allows you to share a single storage between services.
And each service which uses own namespace will have own independent space of keys.*/
func (s *Session) SetNamespace(namespace string) {
	cnamespace := C.CString(namespace)
	defer C.free(unsafe.Pointer(cnamespace))
	C.session_set_namespace(s.session, cnamespace, C.int(len(namespace)))
}

/*
	Read
*/

//ReadResult wraps one result of read operation.
type ReadResult interface {
	// server's reply
	Cmd() *DnetCmd

	// server's address
	Addr() *DnetAddr

	// IO parameters for given
	IO() *DnetIOAttr

	//Data returns string represntation of read data
	Data() string

	// read error
	Error() error
}

type readResult struct {
	cmd    DnetCmd
	addr   DnetAddr
	ioattr DnetIOAttr
	data   string
	err    error
}

func (r *readResult) Cmd() *DnetCmd {
	return &r.cmd
}
func (r *readResult) Addr() *DnetAddr {
	return &r.addr
}
func (r *readResult) IO() *DnetIOAttr {
	return &r.ioattr
}
func (r *readResult) Data() string {
	return r.data
}
func (r *readResult) Error() error {
	return r.err
}

//ReadKey performs a read operation by key.
func (s *Session) ReadKey(key *Key) <-chan ReadResult {
	responseCh := make(chan ReadResult, defaultVOLUME)

	keepaliver := make(chan struct{}, 0)

	onResult := func(result readResult) {
		responseCh <- &result
	}

	onFinish := func(err error) {
		if err != nil {
			responseCh <- &readResult{err: err}
		}

		close(responseCh)

		// close keepalive context
		close(keepaliver)
	}

	go func() {
		<-keepaliver
		onResult = nil
		onFinish = nil
	}()

	C.session_read_data(s.session,
		unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
		key.key)
	return responseCh
}

//ReadKey performs a read operation by string representation of key.
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

//Lookuper represents one result of Write and Lookup operations.
type Lookuper interface {
	// server's reply
	Cmd() *DnetCmd

	// server's address
	Addr() *DnetAddr

	// dnet_file_info structure contains basic information about key location
	Info() *DnetFileInfo

	// address of the node which hosts given key
	StorageAddr() *DnetAddr

	//Path returns a path to file hosting given key on the storage.
	Path() string

	//Error returns string respresentation of error.
	Error() error
}

type lookupResult struct {
	cmd          DnetCmd
	addr         DnetAddr
	info         DnetFileInfo
	storage_addr DnetAddr
	path         string
	err          error
}

func (l *lookupResult) Cmd() *DnetCmd {
	return &l.cmd
}
func (l *lookupResult) Addr() *DnetAddr {
	return &l.addr
}
func (l *lookupResult) Info() *DnetFileInfo {
	return &l.info
}
func (l *lookupResult) StorageAddr() *DnetAddr {
	return &l.storage_addr
}
func (l *lookupResult) Path() string {
	return l.path
}
func (l *lookupResult) Error() error {
	return l.err
}

//WriteData writes blob by a given string representation of Key.
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

//WriteKey writes blob by Key.
func (s *Session) WriteKey(key *Key, blob string) <-chan Lookuper {
	responseCh := make(chan Lookuper, defaultVOLUME)
	raw_data := C.CString(blob) // Mustn't call free. Elliptics does it.

	keepaliver := make(chan struct{}, 0)

	onResult := func(lookup *lookupResult) {
		responseCh <- lookup
	}

	onFinish := func(err error) {
		if err != nil {
			responseCh <- &lookupResult{err: err}
		}
		close(responseCh)

		close(keepaliver)
	}

	go func() {
		<-keepaliver
		onResult = nil
		onFinish = nil
	}()

	C.session_write_data(s.session,
		unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
		key.key, raw_data, C.size_t(len(blob)))
	return responseCh
}

//Lookup returns an information about given Key.
func (s *Session) Lookup(key *Key) <-chan Lookuper {
	responseCh := make(chan Lookuper, defaultVOLUME)
	keepaliver := make(chan struct{}, 0)

	onResult := func(lookup *lookupResult) {
		responseCh <- lookup
	}

	onFinish := func(err error) {
		if err != nil {
			responseCh <- &lookupResult{err: err}
		}
		close(responseCh)

		close(keepaliver)
	}

	/* To keep callbacks alive */
	go func() {
		<-keepaliver
		onResult = nil
		onFinish = nil
	}()

	C.session_lookup(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish), key.key)
	return responseCh
}

/*
	Remove
*/

//Remover wraps information about remove operation.
type Remover interface {
	//Error of remove operation.
	Error() error
}

type removeResult struct {
	err error
}

func (r *removeResult) Error() error {
	return r.err
}

//Remove performs remove operation by a string.
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

//RemoveKey performs remove operation by key.
func (s *Session) RemoveKey(key *Key) <-chan Remover {
	responseCh := make(chan Remover, defaultVOLUME)
	keepaliver := make(chan struct{})

	onResult := func() {
		//It's never called.
	}
	onFinish := func(err error) {
		if err != nil {
			responseCh <- &removeResult{err: err}
		}
		close(responseCh)

		close(keepaliver)
	}

	go func() {
		<-keepaliver
		onResult = nil
		onFinish = nil
	}()

	C.session_remove(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish), key.key)
	return responseCh
}

/*
	Find
*/

//Finder is interface to result of find operations with Indexes.
type Finder interface {
	Error() error
	Data() []IndexEntry
}

type findResult struct {
	id   C.struct_dnet_raw_id
	data []IndexEntry
	err  error
}

//IndexEntry represents one result of some index operations.
type IndexEntry struct {
	//Data is information associated with index.
	Data string
	err  error
}

func (f *findResult) Data() []IndexEntry {
	return f.data
}

func (f *findResult) Error() error {
	return f.err
}

//FindAllIndexes returns IndexEntries for keys, which were indexed with all of indexes.
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

//FindAnyIndexes returns IndexEntries for keys, which were indexed with any of indexes.
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

	keepaliver := make(chan struct{})

	_result := func(result *findResult) {
		responseCh <- result
	}
	onResult = unsafe.Pointer(&_result)

	_finish := func(err error) {
		if err != nil {
			responseCh <- &findResult{err: err}
		}
		close(responseCh)

		close(keepaliver)
	}

	go func() {
		<-keepaliver
		_result = nil
		_finish = nil
	}()

	onFinish = unsafe.Pointer(&_finish)
	return
}

/*
	Indexes
*/

//Indexer is an interface to result of any CRUD operations with indexes.
type Indexer interface {
	//Error returns string representation of error.
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

	keepaliver := make(chan struct{})

	onResult := func() {
		//It's never called. For the future.
	}

	onFinish := func(err error) {
		if err != nil {
			responseCh <- &indexResult{err: err}
		}
		close(responseCh)

		close(keepaliver)
	}

	go func() {
		<-keepaliver
		onResult = nil
		onFinish = nil
	}()

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

//SetIndexes sets indexes for a given key.
func (s *Session) SetIndexes(key string, indexes map[string]string) <-chan Indexer {
	return s.setOrUpdateIndexes(indexesSet, key, indexes)
}

//UpdateIndexes sets indexes for a given key.
func (s *Session) UpdateIndexes(key string, indexes map[string]string) <-chan Indexer {
	return s.setOrUpdateIndexes(indexesUpdate, key, indexes)
}

//ListIndexes gets list of all indxes, which are associated with key.
func (s *Session) ListIndexes(key string) <-chan IndexEntry {
	responseCh := make(chan IndexEntry, defaultVOLUME)
	ekey, err := NewKey(key)
	if err != nil {
		panic(err)
	}
	defer ekey.Free()

	keepaliver := make(chan struct{})

	onResult := func(indexentry *IndexEntry) {
		responseCh <- *indexentry
	}

	onFinish := func(err error) {
		if err != nil {
			responseCh <- IndexEntry{err: err}
		}
		close(responseCh)

		close(keepaliver)
	}

	go func() {
		<-keepaliver
		onResult = nil
		onFinish = nil
	}()

	C.session_list_indexes(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish), ekey.key)
	return responseCh
}

//RemoveIndexes removes indexes from a key.
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

	keepaliver := make(chan struct{})

	onResult := func() {
		//It's never called. For the future.
	}

	onFinish := func(err error) {
		if err != nil {
			responseCh <- &indexResult{err: err}
		}
		close(responseCh)

		close(keepaliver)
	}

	go func() {
		<-keepaliver
		onResult = nil
		onFinish = nil
	}()

	C.session_remove_indexes(s.session,
		unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
		ekey.key, (**C.char)(&cindexes[0]), C.size_t(len(cindexes)))
	return responseCh
}
