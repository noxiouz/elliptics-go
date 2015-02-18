/*
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

import (
	"encoding/hex"
	"fmt"
	"unsafe"
)

/*
#include "route.h"
*/
import "C"

type RouteEntry struct {
	id      []byte
	addr    DnetAddr
	group   uint32
	backend int32
}

func (entry *RouteEntry) ID() []byte {
	return entry.id
}

func (r *RouteEntry) String() string {
	return fmt.Sprintf("route entry: %s: group: %d, addr: %s: backend: %d",
		hex.EncodeToString(r.id), r.group, r.addr.String(), r.backend)
}

func NewRouteEntry(entry *C.struct_dnet_route_entry) *RouteEntry {
	// @dnet_route_entry is not packed, so its fields can be accessed directly
	// compare it to @DnetCmd creation which uses special C wrappers to access fields of the packed structure
	return &RouteEntry{
		id:      C.GoBytes(unsafe.Pointer(&entry.id.id[0]), C.int(C.DNET_ID_SIZE)),
		addr:    NewDnetAddr(&entry.addr),
		group:   uint32(entry.group_id),
		backend: int32(entry.backend_id),
	}
}

func (stat *DnetStat) AddRouteEntry(entry *RouteEntry) {
	backend := stat.FindBackend(entry.group, &entry.addr, entry.backend)

	backend.ID = append(backend.ID,
		DnetRawID{
			ID: entry.ID(),
		})
	return
}

//export go_route_callback
func go_route_callback(dnet_entry *C.struct_dnet_route_entry, key uint64) {
	entry := NewRouteEntry(dnet_entry)

	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find session number")
	}

	stat := context.(*DnetStat)
	stat.AddRouteEntry(entry)
	return
}

func (s *Session) GetRoutes(stat *DnetStat) {
	context := NextContext()
	Pool.Store(context, stat)

	C.session_get_routes(s.session, C.context_t(context))
	stat.Finalize()

	Pool.Delete(context)
	return
}
