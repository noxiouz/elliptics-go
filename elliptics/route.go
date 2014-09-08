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
	"encoding/hex"
	"fmt"
	"unsafe"
)

/*
#include "route.h"
#include <stdio.h>
*/
import "C"

type RouteEntry struct {
	ID		[]byte
	Addr		DnetAddr
	Group		uint32
	Backend		uint32
}

func (r *RouteEntry) String() string {
	return fmt.Sprintf("route entry: %s: group: %d, addr: %s: backend: %d",
		hex.EncodeToString(r.ID), r.Group, r.Addr.String(), r.Backend)
}

func NewRouteEntry(entry *C.struct_dnet_route_entry) *RouteEntry {
	return  &RouteEntry {
		ID:		C.GoBytes(unsafe.Pointer(&entry.id.id[0]), C.int(C.DNET_ID_SIZE)),
		Addr:		NewDnetAddr(&entry.addr),
		Group:		uint32(entry.group_id),
		Backend:	uint32(entry.backend_id),
	}
}

//export go_route_callback
func go_route_callback(dnet_entry *C.struct_dnet_route_entry, context unsafe.Pointer) {
	entry := NewRouteEntry(dnet_entry)

	stat := (* DnetStat)(context)
	stat.AddRouteEntry(entry)
	return
}

func (s *Session) GetRoutes(stat *DnetStat) {
	C.session_get_routes(s.session, unsafe.Pointer(stat))
	stat.Finalize()
	return
}
