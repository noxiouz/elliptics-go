/*
 * 2013+ Copyright (c) Anton Tyurin <noxiouz@yandex.ru>
 * 2014+ Copyright (c) Evgeniy Polyakov <zbr@ioremap.net>
 * All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 */

package elliptics

//#include <elliptics/interface.h>
import "C"

type IOflag uint32

const (
	DNET_IO_FLAGS_SKIP_SENDING           = IOflag(C.DNET_IO_FLAGS_SKIP_SENDING)
	DNET_IO_FLAGS_APPEND                 = IOflag(C.DNET_IO_FLAGS_APPEND)
	DNET_IO_FLAGS_PREPARE                = IOflag(C.DNET_IO_FLAGS_PREPARE)
	DNET_IO_FLAGS_COMMIT                 = IOflag(C.DNET_IO_FLAGS_COMMIT)
	DNET_IO_FLAGS_REMOVED                = IOflag(C.DNET_IO_FLAGS_REMOVED)
	DNET_IO_FLAGS_OVERWRITE              = IOflag(C.DNET_IO_FLAGS_OVERWRITE)
	DNET_IO_FLAGS_NOCSUM                 = IOflag(C.DNET_IO_FLAGS_NOCSUM)
	DNET_IO_FLAGS_PLAIN_WRITE            = IOflag(C.DNET_IO_FLAGS_PLAIN_WRITE)
	DNET_IO_FLAGS_NODATA                 = IOflag(C.DNET_IO_FLAGS_NODATA)
	DNET_IO_FLAGS_CACHE                  = IOflag(C.DNET_IO_FLAGS_CACHE)
	DNET_IO_FLAGS_CACHE_ONLY             = IOflag(C.DNET_IO_FLAGS_CACHE_ONLY)
	DNET_IO_FLAGS_CACHE_REMOVE_FROM_DISK = IOflag(C.DNET_IO_FLAGS_CACHE_REMOVE_FROM_DISK)
	DNET_IO_FLAGS_COMPARE_AND_SWAP       = IOflag(C.DNET_IO_FLAGS_COMPARE_AND_SWAP)
	DNET_IO_FLAGS_CHECKSUM               = IOflag(C.DNET_IO_FLAGS_CHECKSUM)
	DNET_IO_FLAGS_WRITE_NO_FILE_INFO     = IOflag(C.DNET_IO_FLAGS_WRITE_NO_FILE_INFO)
)

type Cflag uint64

const (
	DNET_FLAGS_NEED_ACK       = Cflag(C.DNET_FLAGS_NEED_ACK)
	DNET_FLAGS_MORE           = Cflag(C.DNET_FLAGS_MORE)
	DNET_FLAGS_DESTROY        = Cflag(C.DNET_FLAGS_DESTROY)
	DNET_FLAGS_DIRECT         = Cflag(C.DNET_FLAGS_DIRECT)
	DNET_FLAGS_NOLOCK         = Cflag(C.DNET_FLAGS_NOLOCK)
	DNET_FLAGS_CHECKSUM       = Cflag(C.DNET_FLAGS_CHECKSUM)
	DNET_FLAGS_NOCACHE        = Cflag(C.DNET_FLAGS_NOCACHE)
	DNET_FLAGS_DIRECT_BACKEND = Cflag(C.DNET_FLAGS_DIRECT_BACKEND)
	DNET_FLAGS_TRACE_BIT      = Cflag(C.DNET_FLAGS_TRACE_BIT)
	DNET_FLAGS_REPLY          = Cflag(C.DNET_FLAGS_REPLY)
)

type TraceID C.trace_id_t
