/*
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
	"time"
	"unsafe"
)

/*
#include "session.h"
#include <stdio.h>

struct dnet_backend_status_unpacked {
	uint32_t		backend_id;
	int32_t			state;
	uint32_t		defrag_state;
	struct dnet_time	last_start;
	int32_t			last_start_err;
	int			read_only;
	uint32_t		delay;
};

static inline int dnet_backend_status_from_list(struct dnet_backend_status_list *list,
	uint32_t idx, struct dnet_backend_status_unpacked *ret)
{
	struct dnet_backend_status *st;

	if (idx >= list->backends_count) {
		return -1;
	}

	st = &list->backends[idx];

	ret->backend_id = st->backend_id;
	ret->state = st->state;
	ret->defrag_state = st->defrag_state;
	ret->last_start = st->last_start;
	ret->last_start_err = st->last_start_err;
	ret->read_only = (st->read_only != 0);
	ret->delay = st->delay;

	return 0;
}
*/
import "C"

type DnetBackendStatus struct {
	Backend		int32
	State		int32
	DefragState	uint32
	LastStart	time.Time
	LastStartErr	int32
	RO		bool
	Delay		uint32
}

type DnetBackendsStatus struct {
	Backends	[]DnetBackendStatus
	Error		error
}

//export go_backend_status_callback
func go_backend_status_callback(context unsafe.Pointer, list *C.struct_dnet_backend_status_list) {
	res := &DnetBackendsStatus {
		Backends: make([]DnetBackendStatus, 0, list.backends_count),
	}

	for i := 0; i < int(list.backends_count); i++ {
		var st C.struct_dnet_backend_status_unpacked
		if (C.dnet_backend_status_from_list(list, C.uint32_t(i), &st) == 0) {
			bst := DnetBackendStatus {
				Backend:		int32(st.backend_id),
				State:			int32(st.state),
				DefragState:		uint32(st.defrag_state),
				LastStart:		time.Unix(int64(st.last_start.tsec), int64(st.last_start.tnsec)),
				LastStartErr:		int32(st.last_start_err),
				RO:			st.read_only != 0,
				Delay:			uint32(st.delay),
			}

			res.Backends = append(res.Backends, bst)
		}
	}

	ch := *(*chan *DnetBackendsStatus)(context)
	ch <- res
	close(ch)
}

//export go_backend_status_error
func go_backend_status_error(context unsafe.Pointer, cerr *C.struct_go_error) {
	res := &DnetBackendsStatus {
		Error: &DnetError {
			Code:		int(cerr.code),
			Flags:		uint64(cerr.flags),
			Message:	C.GoString(cerr.message),
		},
	}

	ch := *(*chan *DnetBackendsStatus)(context)
	//ch <- res
	close(ch)
}

func (s *Session) BackendsStatus(addr *DnetAddr) <-chan *DnetBackendsStatus {
	ch := make(chan *DnetBackendsStatus, defaultVOLUME)

	var tmp C.struct_dnet_addr
	addr.CAddr(&tmp)
	C.session_backends_status(s.session, &tmp, unsafe.Pointer(&ch));

	return ch
}

func (s *Session) BackendStartDefrag(addr *DnetAddr, backend_id int32) <-chan *DnetBackendsStatus {
	ch := make(chan *DnetBackendsStatus, defaultVOLUME)

	var tmp C.struct_dnet_addr
	addr.CAddr(&tmp)
	C.session_backend_start_defrag(s.session, &tmp, C.uint32_t(backend_id),
		unsafe.Pointer(&ch));

	return ch
}

func (s *Session) BackendEnable(addr *DnetAddr, backend_id int32) <-chan *DnetBackendsStatus {
	ch := make(chan *DnetBackendsStatus, defaultVOLUME)

	var tmp C.struct_dnet_addr
	addr.CAddr(&tmp)
	C.session_backend_enable(s.session, &tmp, C.uint32_t(backend_id),
		unsafe.Pointer(&ch));

	return ch
}

func (s *Session) BackendDisable(addr *DnetAddr, backend_id int32) <-chan *DnetBackendsStatus {
	ch := make(chan *DnetBackendsStatus, defaultVOLUME)

	var tmp C.struct_dnet_addr
	addr.CAddr(&tmp)
	C.session_backend_disable(s.session, &tmp, C.uint32_t(backend_id),
		unsafe.Pointer(&ch));

	return ch
}

func (s *Session) BackendMakeWritable(addr *DnetAddr, backend_id int32) <-chan *DnetBackendsStatus {
	ch := make(chan *DnetBackendsStatus, defaultVOLUME)

	var tmp C.struct_dnet_addr
	addr.CAddr(&tmp)
	C.session_backend_make_writable(s.session, &tmp, C.uint32_t(backend_id),
		unsafe.Pointer(&ch));

	return ch
}

func (s *Session) BackendMakeReadOnly(addr *DnetAddr, backend_id int32) <-chan *DnetBackendsStatus {
	ch := make(chan *DnetBackendsStatus, defaultVOLUME)

	var tmp C.struct_dnet_addr
	addr.CAddr(&tmp)
	C.session_backend_make_readonly(s.session, &tmp, C.uint32_t(backend_id),
		unsafe.Pointer(&ch));

	return ch
}

func (s *Session) BackendSetDelay(addr *DnetAddr, backend_id int32, delay uint32) <-chan *DnetBackendsStatus {
	ch := make(chan *DnetBackendsStatus, defaultVOLUME)

	var tmp C.struct_dnet_addr
	addr.CAddr(&tmp)
	C.session_backend_set_delay(s.session, &tmp, C.uint32_t(backend_id), C.uint32_t(delay),
		unsafe.Pointer(&ch));

	return ch
}
