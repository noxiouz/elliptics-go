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

/*
#include "session.h"
#include <stdio.h>

static uint64_t dnet_cmd_get_trace_id(struct dnet_cmd* d) {
	return d->trace_id;
}

static uint64_t dnet_cmd_get_flags(struct dnet_cmd* d) {
	return d->flags;
}

static uint64_t dnet_cmd_get_size(struct dnet_cmd* d) {
	return d->size;
}

static uint64_t dnet_cmd_get_trans(struct dnet_cmd* d) {
	return d->trans;
}

static uint32_t dnet_cmd_get_group(struct dnet_cmd* d) {
	return d->id.group_id;
}

static int dnet_cmd_get_status(struct dnet_cmd* d) {
	return d->status;
}

static int dnet_cmd_get_cmd(struct dnet_cmd* d) {
	return d->cmd;
}

static int dnet_cmd_get_backend_id(struct dnet_cmd* d) {
	return d->backend_id;
}

*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

const DnetAddrSize int = 32
type DnetAddr struct {
	Addr	[DnetAddrSize]byte
	Family	uint16
}

func NewDnetAddr(addr *C.struct_dnet_addr) DnetAddr {
	a := DnetAddr {
		Family:	uint16(addr.family),
	}

	copy(a.Addr[:], C.GoBytes(unsafe.Pointer(&addr.addr[0]), C.int(addr.addr_len)))
	return a
}

func (a *DnetAddr) String() string {
	var tmp C.struct_dnet_addr
	var arrayptr = uintptr(unsafe.Pointer(&tmp.addr[0]))
	for i := 0; i < len(a.Addr); i++ {
		*(*C.uint8_t)(unsafe.Pointer(arrayptr)) = C.uint8_t(a.Addr[i])
		arrayptr++
	}

	tmp.addr_len = C.uint16_t(len(a.Addr))
	tmp.family = C.uint16_t(a.Family)

	return fmt.Sprintf("%s:%d", C.GoString(C.dnet_server_convert_dnet_addr(&tmp)), a.Family)
}


type DnetRawID struct {
	ID	[]byte
}

type DnetID struct {
	ID	[]byte
	Group	uint32
}

type DnetCmd struct {
	ID	DnetID
	Status	int32
	Cmd	int32
	Backend	int32
	Trace	uint64
	Flags	uint64
	Trans	uint64
	Size	uint64
}

func NewDnetCmd(cmd *C.struct_dnet_cmd) DnetCmd {
	return DnetCmd {
		ID: DnetID {
			ID: C.GoBytes(unsafe.Pointer(&cmd.id.id[0]), C.int(C.DNET_ID_SIZE)),
			Group: uint32(C.dnet_cmd_get_group(cmd)),
		},

		Status:	int32(C.dnet_cmd_get_status(cmd)),
		Cmd:	int32(C.dnet_cmd_get_cmd(cmd)),
		Backend:	int32(C.dnet_cmd_get_backend_id(cmd)),
		Trace:	uint64(C.dnet_cmd_get_trace_id(cmd)),
		Flags:	uint64(C.dnet_cmd_get_flags(cmd)),
		Trans:	uint64(C.dnet_cmd_get_trans(cmd)),
		Size:	uint64(C.dnet_cmd_get_size(cmd)),
	}
}

type DnetIOAttr struct {
	Parent		[]byte
	ID		[]byte

	Start		uint64
	Num		uint64

	Timestamp	time.Time
	UserFlags	uint64

	TotalSize	uint64

	Flags		uint32

	Offset		uint64
	Size		uint64
}

func NewDnetIOAttr(io *C.struct_dnet_io_attr) DnetIOAttr {
	return DnetIOAttr {
		Parent:		C.GoBytes(unsafe.Pointer(&io.parent[0]), C.int(C.DNET_ID_SIZE)),
		ID:		C.GoBytes(unsafe.Pointer(&io.id[0]), C.int(C.DNET_ID_SIZE)),
		Start:		uint64(io.start),
		Num:		uint64(io.num),
		Timestamp:	time.Unix(int64(io.timestamp.tsec), int64(io.timestamp.tnsec)),
		UserFlags:	uint64(io.user_flags),
		TotalSize:	uint64(io.total_size),
		Flags:		uint32(io.flags),
		Offset:		uint64(io.offset),
		Size:		uint64(io.size),
	}
}

type DnetFileInfo struct {
	Csum		[]byte
	Offset		uint64
	Size		uint64
	Mtime		time.Time
}

func NewDnetFileInfo(info *C.struct_dnet_file_info) DnetFileInfo {
	return DnetFileInfo {
		Csum:	C.GoBytes(unsafe.Pointer(&info.checksum[0]), C.int(C.DNET_ID_SIZE)),
		Offset:		uint64(info.offset),
		Size:		uint64(info.size),
		Mtime:		time.Unix(int64(info.mtime.tsec), int64(info.mtime.tnsec)),
	}
}
