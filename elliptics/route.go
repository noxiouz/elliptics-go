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
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
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

func NewRouteEntry(entry *C.struct_dnet_route_entry) RouteEntry {
	return  RouteEntry {
		ID:		C.GoBytes(unsafe.Pointer(&entry.id.id[0]), C.int(C.DNET_ID_SIZE)),
		Addr:		NewDnetAddr(&entry.addr),
		Group:		uint32(entry.group_id),
		Backend:	uint32(entry.backend_id),
	}
}

type AddressBackend struct {
	Addr		DnetAddr
	Backend		uint32
}

func (ab *AddressBackend) String() string {
	return fmt.Sprintf("ab: address: %s, backend: %d", ab.Addr.String(), ab.Backend)
}

type IDArray struct {
	// @sum hosts total sum of uint64 ids created frm DnetRawID
	// after it is devided by math.MaxUint64 it becomes @Percentage
	sum		uint64

	// percentage of the whole IDs ring currently occupied by given ids
	// and thus node (address + backend)
	Percentage	float64

	// All range starts (IDs) for given node (server address + backend)
	ID		[]DnetRawID
}

func (ida *IDArray) Len() int {
	return len(ida.ID)
}
func (ida *IDArray) Swap(i, j int) {
	ida.ID[i], ida.ID[j] = ida.ID[j], ida.ID[i]
}
func (ida *IDArray) Less(i, j int) bool {
	return bytes.Compare(ida.ID[i].ID, ida.ID[j].ID) < 0
}

func (ids *IDArray) String() string {
	tmp := make([]string, 0, len(ids.ID))
	for i := range ids.ID {
		tmp = append(tmp, hex.EncodeToString(ids.ID[i].ID)[:12])
	}

	return strings.Join(tmp, " ")
}

func NewIDArray() *IDArray {
	return &IDArray {
		ID:		make([]DnetRawID, 0),
		Percentage:	0,
	}
}

type RouteGroup struct {
	Entry		[]RouteEntry
	Ab		map[AddressBackend]*IDArray
}

func (rg *RouteGroup) AddEntry(entry RouteEntry) error {
	rg.Entry = append(rg.Entry, entry)

	ab := AddressBackend {
		Addr:		entry.Addr,
		Backend:	entry.Backend,
	}

	raw := DnetRawID {
		ID: entry.ID,
	}

	ida, ok := rg.Ab[ab]
	if !ok {
		ida = NewIDArray()
		ida.ID = append(ida.ID, raw)
		rg.Ab[ab] = ida
	} else {
		ida.ID = append(ida.ID, raw)
	}

	return nil
}

type slice_uint64 []uint64

func (ids slice_uint64) Len() int {
	return len(ids)
}
func (ids slice_uint64) Swap(i, j int) {
	ids[i], ids[j] = ids[j], ids[i]
}
func (ids slice_uint64) Less(i, j int) bool {
	return ids[i] < ids[j]
}


func (rg *RouteGroup) Finalize() {
	ids2ida := make(map[uint64]*IDArray)
	ids := make([]uint64, 0)

	for _, ida := range rg.Ab {
		sort.Sort(ida)

		for i := range ida.ID {
			val :=	uint64(ida.ID[i].ID[0]) << (7 * 8) |
				uint64(ida.ID[i].ID[1]) << (6 * 8) |
				uint64(ida.ID[i].ID[2]) << (5 * 8) |
				uint64(ida.ID[i].ID[3]) << (4 * 8) |
				uint64(ida.ID[i].ID[4]) << (3 * 8) |
				uint64(ida.ID[i].ID[5]) << (2 * 8) |
				uint64(ida.ID[i].ID[6]) << (1 * 8) |
				uint64(ida.ID[i].ID[7]) << (0 * 8);
			_, has := ids2ida[val]
			if !has {
				ids2ida[val] = ida
				ids = append(ids, val)
			}
		}
	}

	sort.Sort(slice_uint64(ids))

	for i := range ids {
		var prev_val uint64 = 0
		var prev_idx int = 0

		if i == 0 {
			prev_idx = len(ids) - 1
			prev_val = 0
		} else {
			prev_idx = i - 1
			prev_val = ids[prev_idx]
		}

		diff := ids[i] - prev_val
		ids2ida[ids[prev_idx]].sum += diff
	}

	if len(ids) > 0 {
		last := len(ids) - 1
		last_id := ids[last]
		ids2ida[last_id].sum += math.MaxUint64 - last_id
	}

	for _, ida := range rg.Ab {
		ida.Percentage = float64(ida.sum) / float64(math.MaxUint64) * 100.0
	}
}

func NewRouteGroup() RouteGroup {
	return RouteGroup {
		Entry:	make([]RouteEntry, 0),
		Ab:	make(map[AddressBackend]*IDArray),
	}
}

type RouteResult interface {
	RouteGroupAllEntries(group uint32) ([]RouteEntry)
	RouteGroup(group uint32, addr DnetAddr, backend uint32) ([]DnetRawID)
}

type routeResult struct {
	Group	map[uint32]*RouteGroup
}

func (r *routeResult) RouteGroupAllEntries(group uint32) ([]RouteEntry) {
	rg, ok := r.Group[group]
	if !ok {
		return nil
	}

	return rg.Entry
}

func (r *routeResult) RouteGroup(group uint32, addr DnetAddr, backend uint32) ([]DnetRawID) {
	rg, ok := r.Group[group]
	if !ok {
		return nil
	}

	ab := AddressBackend {
		Addr:		addr,
		Backend:	backend,
	}

	ida, ok := rg.Ab[ab]
	if !ok {
		return nil
	}

	return ida.ID
}

//export go_route_callback
func go_route_callback(dnet_entry *C.struct_dnet_route_entry, context unsafe.Pointer) {
	var r *routeResult = (* routeResult)(context)
	entry := NewRouteEntry(dnet_entry)

	rg, ok := r.Group[entry.Group]
	if !ok {
		nrg := NewRouteGroup()
		rg = &nrg
		r.Group[entry.Group] = rg
	}

	rg.AddEntry(entry)
	return
}

func (s *Session) GetRoutes() (route RouteResult, err error) {
	var r routeResult
	r.Group = make(map[uint32]*RouteGroup)
	C.session_get_routes(s.session, unsafe.Pointer(&r))

	for group, rg := range r.Group {
		rg.Finalize()

		fmt.Printf("group: %d:\n", group)
		for ab, ids := range rg.Ab {
			fmt.Printf("%s: ids: %d: %f\n", ab.String(), len(ids.ID), ids.Percentage)
		}
	}

	return &r, nil
}
