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
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"
	//"unsafe"
)

/*
//#include "stat.h"
#include <stdio.h>
*/
import "C"


type AddressBackend struct {
	Addr		DnetAddr
	Backend		uint32
}

func (ab *AddressBackend) String() string {
	return fmt.Sprintf("ab: address: %s, backend: %d", ab.Addr.String(), ab.Backend)
}

type StatBackend struct {
	// All range starts (IDs) for given node (server address + backend)
	ID		[]DnetRawID

	// @sum hosts total sum of uint64 ids created frm DnetRawID
	// after it is devided by math.MaxUint64 it becomes @Percentage
	sum		uint64

	// percentage of the whole IDs ring currently occupied by given ids
	// and thus node (address + backend)
	Percentage	float64
}
func NewStatBackend() *StatBackend {
	return &StatBackend {
		ID:		make([]DnetRawID, 0),
		Percentage:	0,
		sum:		0,
	}
}

func (backend *StatBackend) StatBackendData() (reply map[string]interface{}) {
	reply = make(map[string]interface{})

	reply["percentage"] = backend.Percentage
	return reply
}


func (sb *StatBackend) Len() int {
	return len(sb.ID)
}
func (sb *StatBackend) Swap(i, j int) {
	sb.ID[i], sb.ID[j] = sb.ID[j], sb.ID[i]
}
func (sb *StatBackend) Less(i, j int) bool {
	return bytes.Compare(sb.ID[i].ID, sb.ID[j].ID) < 0
}

func (sb *StatBackend) IDString() string {
	tmp := make([]string, 0, len(sb.ID))
	for i := range sb.ID {
		tmp = append(tmp, hex.EncodeToString(sb.ID[i].ID)[:12])
	}

	return strings.Join(tmp, " ")
}

// @StatGroup hosts mapping from node address + backend into per-backend statistics
// Every group in elliptics contains one or more servers each of which contains one or more backends
// Basically, the lowest IO entitiy is backend which is tightly bound with server node (or node's address)
type StatGroup struct {
	Ab		map[AddressBackend]*StatBackend
}
func NewStatGroup() StatGroup {
	return StatGroup {
		Ab:	make(map[AddressBackend]*StatBackend),
	}
}

func (sg *StatGroup) StatGroupData() (reply []interface{}) {
	reply = make([]interface{}, 0, len(sg.Ab))

	for ab, backend := range sg.Ab {
		tmp := struct {
			Address		string
			Backend		uint32
			Stat		interface{}
		} {
			Address:	ab.Addr.String(),
			Backend:	ab.Backend,
			Stat:		backend.StatBackendData(),
		}
		reply = append(reply, tmp)
	}

	return reply
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

func (rg *StatGroup) Finalize() {
	ids2backend := make(map[uint64]*StatBackend)
	ids := make([]uint64, 0)

	for _, backend := range rg.Ab {
		sort.Sort(backend)

		for i := range backend.ID {
			val :=	uint64(backend.ID[i].ID[0]) << (7 * 8) |
				uint64(backend.ID[i].ID[1]) << (6 * 8) |
				uint64(backend.ID[i].ID[2]) << (5 * 8) |
				uint64(backend.ID[i].ID[3]) << (4 * 8) |
				uint64(backend.ID[i].ID[4]) << (3 * 8) |
				uint64(backend.ID[i].ID[5]) << (2 * 8) |
				uint64(backend.ID[i].ID[6]) << (1 * 8) |
				uint64(backend.ID[i].ID[7]) << (0 * 8);
			_, has := ids2backend[val]
			if !has {
				ids2backend[val] = backend
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
		ids2backend[ids[prev_idx]].sum += diff
	}

	if len(ids) > 0 {
		last := len(ids) - 1
		last_id := ids[last]
		ids2backend[last_id].sum += math.MaxUint64 - last_id
	}

	for _, backend := range rg.Ab {
		backend.Percentage = float64(backend.sum) / float64(math.MaxUint64)
	}
}

type DnetStat struct {
	Group	map[uint32]*StatGroup
}

func (session *Session) DnetStat() *DnetStat {
	st := &DnetStat {
		Group:		make(map[uint32]*StatGroup),
	}

	session.GetRoutes(st)
	return st
}

func (stat *DnetStat) AddRouteEntry(entry *RouteEntry) {
	sg, ok := stat.Group[entry.Group]
	if !ok {
		nsg := NewStatGroup()
		sg = &nsg
		stat.Group[entry.Group] = sg
	}

	ab := AddressBackend {
		Addr:		entry.Addr,
		Backend:	entry.Backend,
	}

	raw := DnetRawID {
		ID: entry.ID,
	}

	backend, ok := sg.Ab[ab]
	if !ok {
		backend = NewStatBackend()
		sg.Ab[ab] = backend
	}

	backend.ID = append(backend.ID, raw)

	return
}

func (stat *DnetStat) Finalize() {
	for group, sg := range stat.Group {
		sg.Finalize()

		fmt.Printf("group: %d:\n", group)
		for ab, ids := range sg.Ab {
			fmt.Printf("%s: ids: %d: %f\n", ab.String(), len(ids.ID), ids.Percentage)
		}
	}
}

func (stat *DnetStat) StatBackend(group uint32, addr DnetAddr, backend_id uint32) *StatBackend {
	sg, ok := stat.Group[group]
	if !ok {
		return nil
	}

	ab := AddressBackend {
		Addr:		addr,
		Backend:	backend_id,
	}

	backend, ok := sg.Ab[ab]
	if !ok {
		return nil
	}

	return backend
}

func (stat *DnetStat) StatData() (reply map[string]interface{}) {
	reply = make(map[string]interface{})
	for group, sg := range stat.Group {
		reply[fmt.Sprintf("%d", group)] = sg.StatGroupData()
	}
	return reply
}
