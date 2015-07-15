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

//=======================================================

static uint64_t dnet_cmd_get_trace_id_raw(uintptr_t d) {
	return ((struct go_read_result*)d)->cmd->trace_id;
}

static uint64_t dnet_cmd_get_flags_raw(uintptr_t d) {
	return ((struct go_read_result*)d)->cmd->flags;
}

static uint64_t dnet_cmd_get_size_raw(uintptr_t d) {
	return ((struct go_read_result*)d)->cmd->size;
}

static uint64_t dnet_cmd_get_trans_raw(uintptr_t d) {
	return ((struct go_read_result*)d)->cmd->trans;
}

static uint32_t dnet_cmd_get_group_raw(uintptr_t d) {
	return ((struct go_read_result*)d)->cmd->id.group_id;
}

static int dnet_cmd_get_status_raw(uintptr_t d) {
	return ((struct go_read_result*)d)->cmd->status;
}

static int dnet_cmd_get_cmd_raw(uintptr_t d) {
	return ((struct go_read_result*)d)->cmd->cmd;
}

static int dnet_cmd_get_backend_id_raw(uintptr_t d) {
	return ((struct go_read_result*)d)->cmd->backend_id;
}


static const void* dnet_cmd_get_id_id_raw(uintptr_t d) {
	return ((struct go_read_result*)d)->cmd->id.id;
}

//=======================================================

static uint64_t dnet_io_attr_get_start(struct dnet_io_attr *io) {
	return io->start;
}
static uint64_t dnet_io_attr_get_num(struct dnet_io_attr *io) {
	return io->num;
}
static int64_t dnet_io_attr_get_tsec(struct dnet_io_attr *io) {
	return io->timestamp.tsec;
}
static int64_t dnet_io_attr_get_tnsec(struct dnet_io_attr *io) {
	return io->timestamp.tsec;
}
static uint64_t dnet_io_attr_get_user_flags(struct dnet_io_attr *io) {
	return io->user_flags;
}
static uint64_t dnet_io_attr_get_total_size(struct dnet_io_attr *io) {
	return io->total_size;
}
static uint64_t dnet_io_attr_get_flags(struct dnet_io_attr *io) {
	return io->flags;
}
static uint64_t dnet_io_attr_get_offset(struct dnet_io_attr *io) {
	return io->offset;
}
static uint64_t dnet_io_attr_get_size(struct dnet_io_attr *io) {
	return io->size;
}

//==========================================================
static uint64_t dnet_io_attr_get_start_raw(uintptr_t io) {
	return ((struct go_read_result *)io)->io_attribute->start;
}
static uint64_t dnet_io_attr_get_num_raw(uintptr_t io) {
	return ((struct go_read_result *)io)->io_attribute->num;
}
static int64_t dnet_io_attr_get_tsec_raw(uintptr_t io) {
	return ((struct go_read_result *)io)->io_attribute->timestamp.tsec;
}
static int64_t dnet_io_attr_get_tnsec_raw(uintptr_t io) {
	return ((struct go_read_result *)io)->io_attribute->timestamp.tsec;
}
static uint64_t dnet_io_attr_get_user_flags_raw(uintptr_t io) {
	return ((struct go_read_result *)io)->io_attribute->user_flags;
}
static uint64_t dnet_io_attr_get_total_size_raw(uintptr_t io) {
	return ((struct go_read_result *)io)->io_attribute->total_size;
}
static uint64_t dnet_io_attr_get_flags_raw(uintptr_t io) {
	return ((struct go_read_result *)io)->io_attribute->flags;
}
static uint64_t dnet_io_attr_get_offset_raw(uintptr_t io) {
	return ((struct go_read_result *)io)->io_attribute->offset;
}
static uint64_t dnet_io_attr_get_size_raw(uintptr_t io) {
	return ((struct go_read_result *)io)->io_attribute->size;
}

static const void* dnet_io_attr_get_parent_raw(uintptr_t io) {
	return &((struct go_read_result *)io)->io_attribute->parent;
}

static const void* dnet_io_attr_get_id_raw(uintptr_t io) {
	return &((struct go_read_result *)io)->io_attribute->id;
}

//===============Addr=====
static uint16_t dnet_addr_get_family(uintptr_t addr) {
	return ((struct go_read_result *)addr)->addr->family;
}

static int dnet_addr_len(uintptr_t addr) {
	return ((struct go_read_result *)addr)->addr->addr_len;
}

static const void* dnet_addr_get_addr(uintptr_t addr) {
	return &((struct go_read_result *)addr)->addr->addr;
}

//=============
#include <stdio.h>

static uint64_t dnet_read_result_size(uintptr_t result) {
	return ((struct go_read_result*)result)->size;
}

static const char* dnet_read_result_file(uintptr_t result) {
	return ((struct go_read_result*)result)->file;
}

*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

type DnetAddr struct {
	Addr   []byte
	Family uint16
}

func NewDnetFile(uresult uintptr) []byte {
	result := C.uintptr_t(uresult)

	size := C.dnet_read_result_size(result)
	if size > 0 {
		return []byte(C.GoStringN(C.dnet_read_result_file(result), C.int(size)))
	} else {
		return make([]byte, 0)
	}
}

func NewDnetAddr(addr *C.struct_dnet_addr) DnetAddr {
	return DnetAddr{
		Family: uint16(addr.family),
		Addr:   C.GoBytes(unsafe.Pointer(&addr.addr[0]), C.int(addr.addr_len)),
	}
}

func NewDnetAddrRaw(uaddr uintptr) DnetAddr {
	addr := C.uintptr_t(uaddr)
	return DnetAddr{
		Family: uint16(C.dnet_addr_get_family(addr)),
		Addr:   C.GoBytes(C.dnet_addr_get_addr(addr), C.dnet_addr_len(addr)),
	}
}

func NewDnetAddrStr(addr_str string) (DnetAddr, error) {
	var caddr *C.struct_dnet_addr = C.dnet_addr_alloc()
	defer C.dnet_addr_free(caddr)

	caddr_str := C.CString(addr_str)
	defer C.free(unsafe.Pointer(caddr_str))

	err := int(C.dnet_create_addr_str(caddr, caddr_str, C.int(len(addr_str))))
	if err < 0 {
		return DnetAddr{}, fmt.Errorf("could not create addr '%s': %d", addr_str, err)
	}

	return NewDnetAddr(caddr), nil
}

func (a *DnetAddr) CAddr(tmp *C.struct_dnet_addr) {
	length := len(a.Addr)
	if length > int(C.DNET_ADDR_SIZE) {
		length = int(C.DNET_ADDR_SIZE)
	}
	if length > 0 {
		C.memcpy(unsafe.Pointer(&tmp.addr[0]), unsafe.Pointer(&a.Addr[0]), C.size_t(length))
	}
	tmp.addr_len = C.uint16_t(length)
	tmp.family = C.uint16_t(a.Family)
}

func (a *DnetAddr) String() string {
	var tmp *C.struct_dnet_addr = C.dnet_addr_alloc()
	defer C.dnet_addr_free(tmp)

	a.CAddr(tmp)
	return fmt.Sprintf("%s:%d", C.GoString(C.dnet_addr_string(tmp)), a.Family)
}

func (a *DnetAddr) HostString() string {
	var tmp *C.struct_dnet_addr = C.dnet_addr_alloc()
	defer C.dnet_addr_free(tmp)

	a.CAddr(tmp)
	return fmt.Sprintf("%s", C.GoString(C.dnet_addr_host_string(tmp)))
}

type DnetRawID struct {
	ID []byte
}

type DnetID struct {
	ID    []byte
	Group uint32
}

type DnetCmd struct {
	ID      DnetID
	Status  int32
	Cmd     int32
	Backend int32
	Trace   uint64
	Flags   uint64
	Trans   uint64
	Size    uint64
}

func NewDnetCmd(cmd *C.struct_dnet_cmd) DnetCmd {
	return DnetCmd{
		ID: DnetID{
			ID:    C.GoBytes(unsafe.Pointer(&cmd.id.id[0]), C.int(C.DNET_ID_SIZE)),
			Group: uint32(C.dnet_cmd_get_group(cmd)),
		},

		Status:  int32(C.dnet_cmd_get_status(cmd)),
		Cmd:     int32(C.dnet_cmd_get_cmd(cmd)),
		Backend: int32(C.dnet_cmd_get_backend_id(cmd)),
		Trace:   uint64(C.dnet_cmd_get_trace_id(cmd)),
		Flags:   uint64(C.dnet_cmd_get_flags(cmd)),
		Trans:   uint64(C.dnet_cmd_get_trans(cmd)),
		Size:    uint64(C.dnet_cmd_get_size(cmd)),
	}
}

func NewDnetCmdRaw(ucmd uintptr) DnetCmd {
	cmd := C.uintptr_t(ucmd)
	return DnetCmd{
		ID: DnetID{
			ID:    C.GoBytes(C.dnet_cmd_get_id_id_raw(cmd), C.int(C.DNET_ID_SIZE)),
			Group: uint32(C.dnet_cmd_get_group_raw(cmd)),
		},

		Status:  int32(C.dnet_cmd_get_status_raw(cmd)),
		Cmd:     int32(C.dnet_cmd_get_cmd_raw(cmd)),
		Backend: int32(C.dnet_cmd_get_backend_id_raw(cmd)),
		Trace:   uint64(C.dnet_cmd_get_trace_id_raw(cmd)),
		Flags:   uint64(C.dnet_cmd_get_flags_raw(cmd)),
		Trans:   uint64(C.dnet_cmd_get_trans_raw(cmd)),
		Size:    uint64(C.dnet_cmd_get_size_raw(cmd)),
	}
}

type DnetIOAttr struct {
	Parent []byte
	ID     []byte

	Start uint64
	Num   uint64

	Timestamp time.Time
	UserFlags uint64

	TotalSize uint64

	Flags uint32

	Offset uint64
	Size   uint64
}

func NewDnetIOAttrRaw(uio uintptr) DnetIOAttr {
	io := C.uintptr_t(uio)
	return DnetIOAttr{
		Parent:    C.GoBytes(C.dnet_io_attr_get_parent_raw(io), C.int(C.DNET_ID_SIZE)),
		ID:        C.GoBytes(C.dnet_io_attr_get_id_raw(io), C.int(C.DNET_ID_SIZE)),
		Start:     uint64(C.dnet_io_attr_get_start_raw(io)),
		Num:       uint64(C.dnet_io_attr_get_num_raw(io)),
		Timestamp: time.Unix(int64(C.dnet_io_attr_get_tsec_raw(io)), int64(C.dnet_io_attr_get_tnsec_raw(io))),
		UserFlags: uint64(C.dnet_io_attr_get_user_flags_raw(io)),
		TotalSize: uint64(C.dnet_io_attr_get_total_size_raw(io)),
		Flags:     uint32(C.dnet_io_attr_get_flags_raw(io)),
		Offset:    uint64(C.dnet_io_attr_get_offset_raw(io)),
		Size:      uint64(C.dnet_io_attr_get_size_raw(io)),
	}
}

func NewDnetIOAttr(io *C.struct_dnet_io_attr) DnetIOAttr {
	return DnetIOAttr{
		Parent:    C.GoBytes(unsafe.Pointer(&io.parent[0]), C.int(C.DNET_ID_SIZE)),
		ID:        C.GoBytes(unsafe.Pointer(&io.id[0]), C.int(C.DNET_ID_SIZE)),
		Start:     uint64(C.dnet_io_attr_get_start(io)),
		Num:       uint64(C.dnet_io_attr_get_num(io)),
		Timestamp: time.Unix(int64(C.dnet_io_attr_get_tsec(io)), int64(C.dnet_io_attr_get_tnsec(io))),
		UserFlags: uint64(C.dnet_io_attr_get_user_flags(io)),
		TotalSize: uint64(C.dnet_io_attr_get_total_size(io)),
		Flags:     uint32(C.dnet_io_attr_get_flags(io)),
		Offset:    uint64(C.dnet_io_attr_get_offset(io)),
		Size:      uint64(C.dnet_io_attr_get_size(io)),
	}
}

type DnetFileInfo struct {
	Csum   []byte
	Offset uint64
	Size   uint64
	Mtime  time.Time
}

func NewDnetFileInfo(info *C.struct_dnet_file_info) DnetFileInfo {
	return DnetFileInfo{
		Csum:   C.GoBytes(unsafe.Pointer(&info.checksum[0]), C.int(C.DNET_ID_SIZE)),
		Offset: uint64(info.offset),
		Size:   uint64(info.size),
		Mtime:  time.Unix(int64(info.mtime.tsec), int64(info.mtime.tnsec)),
	}
}
