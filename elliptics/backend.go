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
	"fmt"
	"time"
	"unsafe"
)

/*
#include "session.h"
#include <stdio.h>
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
	backends	[]DnetBackendStatus
	err		error
}

//export go_backend_status_callback
func go_backend_status_callback(context unsafe.Pointer, elements *C.struct_dnet_backend_status_list) {
	num := elements.backends_count
	cback := &elements.backends

	fmt.Printf("status: %d vs %d backends\n", num, len(cback))

	res := &DnetBackendsStatus {
		backends: make([]DnetBackendStatus, 0, num),
	}


	ch := *(*chan *DnetBackendsStatus)(context)
	ch <- res
	close(ch)
}

//export go_backend_status_error
func go_backend_status_error(context unsafe.Pointer, cerr *C.struct_go_error) {
	res := &DnetBackendsStatus {
		err: &DnetError {
			Code:		int(cerr.code),
			Flags:		uint64(cerr.flags),
			Message:	C.GoString(cerr.message),
		},
	}

	ch := *(*chan *DnetBackendsStatus)(context)
	ch <- res
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
