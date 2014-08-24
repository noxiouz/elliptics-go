/*
 * 2013+ Copyright (c) Anton Tyurin <noxiouz@yandex.ru>
 * 2014+ Copyright (c) Evgeniy Polyakov <zbr@ioremap.net>
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
	"io"
	"io/ioutil"
	"net/http"
	"unsafe"
)

/*
#include "session.h"
#include <stdio.h>
*/
import "C"

const defaultVOLUME = 10
const max_chunk_size uint64 = 10 * 1024 * 1024

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
	groups		[]int32
	session		unsafe.Pointer
}

//NewSession returns Session connected with given Node.
func NewSession(node *Node) (*Session, error) {
	session, err := C.new_elliptics_session(node.node)
	if err != nil {
		return nil, err
	}
	return &Session{
		session: session,
		groups: make([]int32, 0, 0),
	}, err
}

//SetGroups points groups Session should work with.
func (s *Session) SetGroups(groups []int32) {
	C.session_set_groups(s.session, (*C.int32_t)(&groups[0]), C.int(len(groups)))
	s.groups = groups
}
//GetGroups returns array of groups this session holds
func (s *Session) GetGroups() []int32 {
	return s.groups
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
	Data() []byte

	// read error
	Error() error
}

type readResult struct {
	cmd    DnetCmd
	addr   DnetAddr
	ioattr DnetIOAttr
	data   []byte
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
func (r *readResult) Data() []byte {
	return r.data
}
func (r *readResult) Error() error {
	return r.err
}

//ReadKey performs a read operation by key.
func (s *Session) ReadChunk(key *Key, offset, size uint64) <-chan ReadResult {
	responseCh := make(chan ReadResult, defaultVOLUME)
	keepaliver := make(chan struct{}, 0)

	var onResult func(result readResult)
	var onFinish func(err error)

	try_next := func() {
		chunk_size := size
		if chunk_size > max_chunk_size {
			chunk_size = max_chunk_size
		}

		size -= chunk_size
		offset += chunk_size

		C.session_read_data(s.session,
			unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
			key.key, C.uint64_t(offset - chunk_size), C.uint64_t(chunk_size))
	}


	onResult = func(result readResult) {
		responseCh <- &result
	}

	onFinish = func(err error) {
		if err != nil {
			responseCh <- &readResult{err: err}
			close(responseCh)
			close(keepaliver)
			return
		}

		if (size == 0) {
			close(responseCh)
			close(keepaliver)
			return
		}

		try_next()
	}

	go func() {
		<-keepaliver
		onResult = nil
		onFinish = nil
	}()

	try_next()
	return responseCh
}

//StreamData sends a stream read from elliptics into given http response writer
// It doesn't start reading next chunk (10M) until the one already read has not been written
// into the client's pipe. This eliminates number of unneeded copies and adds flow control
// of the client's pips.
func (s *Session) StreamHTTP(kstr string, offset, size uint64, w http.ResponseWriter) error {
	key, err := NewKey(kstr)
	if err != nil {
		return err
	}
	defer key.Free()

	orig_offset := offset
	orig_size := size

	// size == 0 means 'read everything
	for size >= 0 {
		chunk_size := size
		if chunk_size > max_chunk_size || chunk_size == 0 {
			chunk_size = max_chunk_size
		}

		err = &DnetError {
			Code: -6,
			Flags: 0,
			Message: fmt.Sprintf("could not read anything at all"),
		}

		for rd := range s.ReadChunk(key, offset, chunk_size) {
			err = rd.Error()
			if err != nil {
				continue
			}

			if offset == orig_offset {
				w.Header().Set("Content-Length", fmt.Sprintf("%d", rd.IO().TotalSize - offset))
				if size == 0 || size > rd.IO().TotalSize - offset {
					size = rd.IO().TotalSize - offset
				}

				w.WriteHeader(http.StatusOK)
			}

			data := rd.Data()

			w.Write(data)

			offset += uint64(len(data))
			size -= uint64(len(data))
			break
		}

		if err != nil {
			return &DnetError {
				Code: ErrorStatus(err),
				Flags: 0,
				Message: fmt.Sprintf("could not stream data: current-offset: %d/%d, current-size: %d, rest-size: %d/%d: %v",
					orig_offset, offset, chunk_size, orig_size, size, err),
			}
		}

		if size == 0 {
			break
		}
	}

	return nil
}

//ReadKey performs a read operation by key.
func (s *Session) ReadKey(key *Key, offset, size uint64) <-chan ReadResult {
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
		key.key, C.uint64_t(offset), C.uint64_t(size))
	return responseCh
}

//ReadKey performs a read operation by string representation of key.
func (s *Session) ReadData(key string, offset, size uint64) <-chan ReadResult {
	ekey, err := NewKey(key)
	if err != nil {
		errCh := make(chan ReadResult, 1)
		errCh <- &readResult{err: err}
		close(errCh)
		return errCh
	}
	defer ekey.Free()
	return s.ReadKey(ekey, offset, size)
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
func (s *Session) WriteData(key string, input io.Reader, offset, total_size uint64) <-chan Lookuper {
	ekey, err := NewKey(key)
	if err != nil {
		responseCh := make(chan Lookuper, defaultVOLUME)
		responseCh <- &lookupResult{err: err}
		close(responseCh)
		return responseCh
	}
	defer ekey.Free()
	return s.WriteKey(ekey, input, offset, total_size)
}

func (s *Session) WriteChunk(key *Key, input io.Reader, initial_offset, total_size uint64) <-chan Lookuper {
	responseCh := make(chan Lookuper, defaultVOLUME)

	keepaliver := make(chan struct{}, 0)

	chunk := make([]byte, max_chunk_size, max_chunk_size)

	var offset uint64 = initial_offset
	var n64 uint64

	onChunkResult := func(lookup *lookupResult) {
		if total_size == 0 {
			responseCh <- lookup
		}
	}

	var onChunkFinish func(err error)

	onChunkFinish = func(err error) {
		if err != nil {
			responseCh <- &lookupResult{err: err}
			close(responseCh)
			close(keepaliver)
			return
		}

		if total_size == 0 {
			close(responseCh)
			close(keepaliver)
			return
		}

		n, err := input.Read(chunk)
		if n <= 0 && err != nil {
			responseCh <- &lookupResult{err: err}
			close(responseCh)
			close(keepaliver)
			return
		}

		n64 = uint64(n)
		total_size -= n64
		offset += n64

		if total_size != 0 {
			C.session_write_plain(s.session,
				unsafe.Pointer(&onChunkResult), unsafe.Pointer(&onChunkFinish),
				key.key, C.uint64_t(offset - n64),
				(*C.char)(unsafe.Pointer(&chunk[0])), C.uint64_t(n))
		} else {
			C.session_write_commit(s.session,
				unsafe.Pointer(&onChunkResult), unsafe.Pointer(&onChunkFinish),
				key.key, C.uint64_t(offset - n64), C.uint64_t(offset),
				(*C.char)(unsafe.Pointer(&chunk[0])), C.uint64_t(n))
		}
	}

	go func() {
		<-keepaliver
		// this is GC magic - goroutine 'grabs' variables from the WriteKey
		// context, which it used somehow in the code
		// we need to keep all variables that are not copied but referenced
		// in c++ code to be alive long enough to be copied into the socket
		// we keep them referenced until write_data() completes, i.e. onFinish()
		// is called, which in turn writes into @keepalive channel to wake up
		// this goroutine and finish it
		onChunkResult = nil
		onChunkFinish = nil
		_ = chunk
	}()

	rest := total_size
	if rest > max_chunk_size {
		rest = max_chunk_size
	}

	n, err := input.Read(chunk)
	if err != nil {
		responseCh <- &lookupResult{err: err}
		close(responseCh)
		close(keepaliver)
		return responseCh
	}

	n64 = uint64(n)
	total_size -= n64
	offset += n64

	C.session_write_prepare(s.session,
		unsafe.Pointer(&onChunkResult), unsafe.Pointer(&onChunkFinish),
		key.key, C.uint64_t(offset - n64), C.uint64_t(total_size + n64),
		(*C.char)(unsafe.Pointer(&chunk[0])), C.uint64_t(n))

	return responseCh
}

//WriteKey writes blob by Key.
func (s *Session) WriteKey(key *Key, input io.Reader, offset, total_size uint64) <-chan Lookuper {
	if total_size > max_chunk_size {
		return s.WriteChunk(key, input, offset, total_size)
	}

	responseCh := make(chan Lookuper, defaultVOLUME)

	keepaliver := make(chan struct{}, 0)

	onWriteResult := func(lookup *lookupResult) {
		responseCh <- lookup
	}

	onWriteFinish := func(err error) {
		if err != nil {
			responseCh <- &lookupResult{err: err}
		}
		close(responseCh)
		close(keepaliver)
	}

	chunk, err := ioutil.ReadAll(input)
	if err != nil {
		responseCh <- &lookupResult{err: err}
		close(responseCh)
		close(keepaliver)
		return responseCh
	}

	go func() {
		<-keepaliver
		// this is GC magic - goroutine 'grabs' variables from the WriteKey
		// context, which it used somehow in the code
		// we need to keep all variables that are not copied but referenced
		// in c++ code to be alive long enough to be copied into the socket
		// we keep them referenced until write_data() completes, i.e. onFinish()
		// is called, which in turn writes into @keepalive channel to wake up
		// this goroutine and finish it
		onWriteResult = nil
		onWriteFinish = nil
		_ = chunk
	}()

	C.session_write_data(s.session,
		unsafe.Pointer(&onWriteResult), unsafe.Pointer(&onWriteFinish),
		key.key, C.uint64_t(offset), (*C.char)(unsafe.Pointer(&chunk[0])), C.uint64_t(len(chunk)))

	return responseCh
}

// Lookup returns an information about given Key.
// It only returns the first group where key has been found.
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

// ParallelLookup returns all information about given Key,
// it sends multiple lookup requests in parallel to all session groups
// and returns information about all specified group where given key has been found.
func (s *Session) ParallelLookup(kstr string) <-chan Lookuper {
	responseCh := make(chan Lookuper, defaultVOLUME)

	key, err := NewKey(kstr)
	if err != nil {
		responseCh <- &lookupResult{err: err}
		close(responseCh)
		return responseCh
	}
	defer key.Free()

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

	C.session_parallel_lookup(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish), key.key)
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
		(**C.char)(&cindexes[0]), C.uint64_t(len(indexes)))
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
		(**C.char)(&cindexes[0]), C.uint64_t(len(indexes)))
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
			C.uint64_t(len(cindexes)))

	case indexesUpdate:
		C.session_update_indexes(s.session, unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
			ekey.key,
			(**C.char)(&cindexes[0]),
			(*C.struct_go_data_pointer)(&cdatas[0]),
			C.uint64_t(len(cindexes)))
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
		ekey.key, (**C.char)(&cindexes[0]), C.uint64_t(len(cindexes)))
	return responseCh
}
