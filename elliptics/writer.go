package elliptics

import (
	"fmt"
	"unsafe"
	"time"
)

/*
#include "session.h"
#include <stdio.h>
*/
import "C"


// implements Writer and Seeker interfaces
type WriteSeeker struct {
	session			*Session

	key			*Key
	// we can not remove key if it belongs to caller, who has created WriteSeeker object
	// via NewWriteSeekerKey()
	want_key_free		bool

	// how many bytes we want to reserve on disk on remote storage for this key
	reserve_size		uint64

	// how many bytes we want to write into this interface
	total_size		uint64

	// remote offset where to put data
	remote_offset		int64

	// start remote offset where data will be placed, is needed to determine when to call prepare or plain write method
	orig_remote_offset	int64

	// offset within this chunk
	local_offset		int
	chunk			[]byte

	mtime			time.Time
}

func NewWriteSeeker(session *Session, kstr string, remote_offset int64, total_size, reserve_size uint64) (*WriteSeeker, error) {
	key, err := NewKey(kstr)
	if err != nil {
		return nil, err
	}

	ws, err := NewWriteSeekerKey(session, key, remote_offset, total_size, reserve_size)
	if err != nil {
		key.Free()
		return nil, err
	}

	ws.want_key_free = true

	return ws, nil
}

func NewWriteSeekerKey(session *Session, key *Key, remote_offset int64, total_size, reserve_size uint64) (*WriteSeeker, error) {
	w := &WriteSeeker {
		session:		session,
		key:			key,
		want_key_free:		false,
		total_size:		total_size,
		reserve_size:		reserve_size,
		remote_offset:		remote_offset,
		orig_remote_offset:	remote_offset,
		chunk:			make([]byte, 10 * 1024 * 1024),
	}

	return w, nil
}

func (w *WriteSeeker) Free() {
	if w.want_key_free {
		w.key.Free()
	}
}

type write_channel struct {
	response			chan Lookuper
	on_result_context		uint64
	on_finish_context		uint64
}

func new_write_channel() *write_channel {
	wch := &write_channel {
		response:		make(chan Lookuper, 16),
		on_result_context:	NextContext(),
		on_finish_context:	NextContext(),
	}

	Pool.Store(wch.on_result_context, wch.on_result)
	Pool.Store(wch.on_finish_context, wch.on_finish)

	return wch
}

func (wch *write_channel) on_result(lookup *lookupResult) {
	wch.response <- lookup
}

func (wch *write_channel) on_finish(err error) {
	if err != nil {
		wch.response <- &lookupResult{err: err}
	}

	close(wch.response)

	Pool.Delete(wch.on_result_context)
	Pool.Delete(wch.on_finish_context)
}

func (w *WriteSeeker) Flush(buf []byte) (err error) {
	if w.local_offset == 0 {
		return nil
	}

	wch := new_write_channel()

	len64 := uint64(len(buf))

	if (w.remote_offset == w.orig_remote_offset) {
		if uint64(w.remote_offset) + len64 == w.total_size {
			// the whole package fits this chunk, use common write
			C.session_write_data(w.session.session,
				C.context_t(wch.on_result_context), C.context_t(wch.on_finish_context),
				w.key.key, C.uint64_t(w.remote_offset),
				(*C.char)(unsafe.Pointer(&buf[0])), C.uint64_t(len64))
		} else {
			C.session_write_prepare(w.session.session,
				C.context_t(wch.on_result_context), C.context_t(wch.on_finish_context),
				w.key.key, C.uint64_t(w.remote_offset), C.uint64_t(w.reserve_size),
				(*C.char)(unsafe.Pointer(&buf[0])), C.uint64_t(len64))
		}
	} else if uint64(w.remote_offset) + len64 == w.total_size {
		C.session_write_commit(w.session.session,
			C.context_t(wch.on_result_context), C.context_t(wch.on_finish_context),
			w.key.key, C.uint64_t(w.remote_offset), C.uint64_t(w.total_size),
			(*C.char)(unsafe.Pointer(&buf[0])), C.uint64_t(len64))
	} else {
		C.session_write_plain(w.session.session,
			C.context_t(wch.on_result_context), C.context_t(wch.on_finish_context),
			w.key.key, C.uint64_t(w.remote_offset),
			(*C.char)(unsafe.Pointer(&buf[0])), C.uint64_t(len64))
	}

	w.remote_offset += int64(len64)

	have_good_write := false
	errors := make([]error, 0)
	for wr := range wch.response {
		err = wr.Error()
		if err != nil {
			errors = append(errors, err)
			continue
		}

		have_good_write = true
	}

	if have_good_write == false {
		return fmt.Errorf("write failed, remote_offset: %d, reserve_size: %d, total_size: %d, chunk_size: %d, errors: %v",
			w.remote_offset - int64(len64), w.reserve_size, w.total_size, len64, errors)
	}

	return nil
}

func (w *WriteSeeker) Write(p []byte) (int, error) {
	var err error

	data_len := len(p)

	if uint64(w.remote_offset) >= w.total_size {
		return 0, fmt.Errorf("going to write past total_size: %d, reserve_size: %d, remote_offset: %d, chunk_size: %d",
				w.total_size, w.reserve_size, w.remote_offset, len(p))
	}

	// flush internal buffer if new data will overflow it
	if w.local_offset + data_len > len(w.chunk) {
		err = w.Flush(w.chunk[0 : w.local_offset])
		if err != nil {
			return 0, err
		}

		w.local_offset = 0
	}

	// if new data can not be copied into the internal write it directly into elliptics
	if data_len > len(w.chunk) {
		err = w.Flush(p)
		if err != nil {
			return 0, err
		}

		return data_len, nil
	}

	copy(w.chunk[w.local_offset:], p)
	w.local_offset += data_len

	// if we have finished our whole write, i.e. @total_size equals to already written data (@remote_offset) plus
	// what we have in the internal buffer (@w.local_offset), flush internal buffer
	if uint64(w.remote_offset) + uint64(w.local_offset) >= w.total_size {
		err = w.Flush(w.chunk[0 : w.local_offset])
		if err != nil {
			return 0, err
		}

		w.local_offset = 0
	}

	return data_len, nil
}

func (w *WriteSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		w.remote_offset = offset
	case 1:
		w.remote_offset += offset
	case 2:
		w.remote_offset = int64(w.total_size) + offset
	}

	return w.remote_offset, nil
}

func (w *WriteSeeker) SetKey(session *Session, key *Key, remote_offset int64, total_size, reserve_size uint64) error {
	w.Free()

	w.session = session
	w.key = key
	w.want_key_free = false
	w.total_size = total_size
	w.reserve_size = reserve_size

	w.remote_offset = remote_offset
	w.orig_remote_offset = remote_offset

	w.local_offset = 0

	return nil
}
