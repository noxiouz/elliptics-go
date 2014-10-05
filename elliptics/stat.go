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
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"log"
	"math"
	"sort"
	"strings"
	"time"
	"unsafe"
)

/*
#include "stat.h"
*/
import "C"

const (
	StatCategoryCache	int64 =		1 << 0
	StatCategoryIO		int64 =		1 << 1
	StatCategoryCommands	int64 =		1 << 2
	StatCategoryBackend	int64 =		1 << 4
	StatCategoryProcFS	int64 =		1 << 6
)

type AddressBackend struct {
	Addr    DnetAddr
	Backend int32
}

func (ab *AddressBackend) String() string {
	return fmt.Sprintf("ab: address: %s, backend: %d", ab.Addr.String(), ab.Backend)
}

type VFS struct {
	// space in bytes for given backend
	Total, Avail       uint64
	BackendRemovedSize uint64
	BackendUsedSize    uint64
}

type CStat struct {
	RequestsSuccess		uint64
	RequestsFailures	uint64
	Bytes			uint64

	RPSFailures		float64
	RPSSuccess		float64
	BPS			float64
}

type StatBackend struct {
	// All range starts (IDs) for given node (server address + backend)
	ID []DnetRawID

	// @sum hosts total sum of uint64 ids created frm DnetRawID
	// after it is devided by math.MaxUint64 it becomes @Percentage
	sum uint64

	// percentage of the whole IDs ring currently occupied by given ids
	// and thus node (address + backend)
	Percentage float64

	// VFS statistics: available, used and total space
	VFS		VFS

	// per-command size/number counters
	// difference between the two divided by the time difference equals to RPS/BPS
	Commands	map[string]*CStat
}

func NewStatBackend() *StatBackend {
	return &StatBackend {
		ID:		make([]DnetRawID, 0),
		Percentage:	0,
		sum:		0,
		Commands:	make(map[string]*CStat),
	}
}

func (backend *StatBackend) StatBackendData() (reply map[string]interface{}) {
	reply = make(map[string]interface{})

	reply["percentage"] = backend.Percentage
	reply["vfs"] = backend.VFS
	reply["commands"] = backend.Commands
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
	Ab map[AddressBackend]*StatBackend
}

func NewStatGroup() StatGroup {
	return StatGroup{
		Ab: make(map[AddressBackend]*StatBackend),
	}
}

func (sg *StatGroup) StatGroupData() (reply []interface{}) {
	reply = make([]interface{}, 0, len(sg.Ab))

	for ab, backend := range sg.Ab {
		tmp := struct {
			Address string
			Backend int32
			Stat    interface{}
		}{
			Address: ab.Addr.String(),
			Backend: ab.Backend,
			Stat:    backend.StatBackendData(),
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
			val := uint64(backend.ID[i].ID[0])<<(7*8) |
				uint64(backend.ID[i].ID[1])<<(6*8) |
				uint64(backend.ID[i].ID[2])<<(5*8) |
				uint64(backend.ID[i].ID[3])<<(4*8) |
				uint64(backend.ID[i].ID[4])<<(3*8) |
				uint64(backend.ID[i].ID[5])<<(2*8) |
				uint64(backend.ID[i].ID[6])<<(1*8) |
				uint64(backend.ID[i].ID[7])<<(0*8)
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
	Time		time.Time
	Group		map[uint32]*StatGroup
}

type StatEntry struct {
	cmd  DnetCmd
	addr DnetAddr
	stat []byte
	err  error
}

func (entry *StatEntry) Group() uint32 {
	return entry.cmd.ID.Group
}
func (entry *StatEntry) AddressBackend() AddressBackend {
	return AddressBackend{
		Addr:    entry.addr,
		Backend: entry.cmd.Backend,
	}
}

//export go_stat_callback
func go_stat_callback(result *C.struct_go_stat_result, context unsafe.Pointer) {
	callback := *(*func(*StatEntry))(context)

	res := &StatEntry{
		cmd:  NewDnetCmd(result.cmd),
		addr: NewDnetAddr(result.addr),
		stat: C.GoBytes(unsafe.Pointer(result.stat_data), C.int(result.stat_size)),
		err:  nil,
	}

	callback(res)
}

func (s *Session) DnetStat() *DnetStat {
	response := make(chan *StatEntry, 10)
	keepaliver := make(chan struct{}, 0)

	onResult := func(result *StatEntry) {
		response <- result
	}

	onFinish := func(err error) {
		if err != nil {
			response <- &StatEntry{
				err: err,
			}
		}

		close(response)
		close(keepaliver)
	}

	go func() {
		<-keepaliver
		onResult = nil
		onFinish = nil
	}()

	categories := StatCategoryBackend | StatCategoryProcFS | StatCategoryCommands

	C.session_get_stats(s.session,
		unsafe.Pointer(&onResult), unsafe.Pointer(&onFinish),
		C.uint64_t(categories))

	st := &DnetStat{
		Group: make(map[uint32]*StatGroup),
	}

	s.GetRoutes(st)

	// read stat results from the channel and update DnetStat
	for se := range response {
		st.AddStatEntry(se)
	}

	return st
}

// @Diff() updates differential counters like success/failure RPS and BPS
// i.e. those counters which require difference measured for some time
func (stat *DnetStat) Diff(prev *DnetStat) {
	if prev == nil {
		return
	}

	duration := stat.Time.Sub(prev.Time).Seconds()

	for group, sg := range stat.Group {
		psg, ok := prev.Group[group]
		if !ok {
			continue
		}

		for ab, sb := range sg.Ab {
			psb, ok := psg.Ab[ab]
			if !ok {
				continue
			}

			for cmd, cstat := range sb.Commands {
				pcstat, ok := psb.Commands[cmd]
				if !ok {
					continue
				}

				cstat.RPSSuccess = float64(cstat.RequestsSuccess - pcstat.RequestsSuccess) / duration
				cstat.RPSFailures = float64(cstat.RequestsFailures - pcstat.RequestsFailures) / duration
				cstat.BPS = float64(cstat.Bytes - pcstat.Bytes) / duration
			}
		}
	}
}

func (stat *DnetStat) FindBackend(group uint32, addr *DnetAddr, backend_id int32) *StatBackend {
	sg, ok := stat.Group[group]
	if !ok {
		nsg := NewStatGroup()
		sg = &nsg
		stat.Group[group] = sg
	}

	ab := AddressBackend{
		Addr:    *addr,
		Backend: backend_id,
	}

	backend, ok := sg.Ab[ab]
	if !ok {
		backend = NewStatBackend()
		sg.Ab[ab] = backend
	}

	return backend
}

func (stat *DnetStat) AddStatEntry(entry *StatEntry) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Shitty json: %v\n", r)
		}
	}()

	const (
		BackendStateDisabled     int = 0
		BackendStateEnabled      int = 1
		BackendStateActivating   int = 2
		BackendStateDeactivating int = 3

		DefragStateNotStarted int = 0
		DefragStateInProgress int = 1
	)

	var (
		backend_state = map[int]string{
			BackendStateDisabled:     "disabled",
			BackendStateEnabled:      "enabled",
			BackendStateActivating:   "activating",
			BackendStateDeactivating: "deactivating",
		}
		defrag_state = map[int]string{
			DefragStateNotStarted: "not-started",
			DefragStateInProgress: "in-progress",
		}
		_ = defrag_state
		_ = backend_state
	)

	var r Response

	err := json.Unmarshal(entry.stat, &r)
	if err != nil {
		log.Printf("%s: could not parse stat entry '%s' reply: %v\n", entry.addr.String(), string(entry.stat), err)
	}

	stat.Time = time.Unix(int64(r.Timestamp.Sec), int64(r.Timestamp.USec * 1000))

	if r.MonitorStatus != "enabled" {
		log.Printf("%s: monitoring doesn't work: %v\n", entry.addr.String(), r.MonitorStatus)
		return
	}

	for _, vnode := range r.Backends {
		if vnode.Status.State == BackendStateEnabled {
			backend := stat.FindBackend(vnode.Backend.Config.Group, &entry.addr, int32(vnode.BackendID))
			backend.VFS.Total = vnode.Backend.VFS.FrSize * vnode.Backend.VFS.Blocks
			backend.VFS.Avail = vnode.Backend.VFS.BFree * vnode.Backend.VFS.BSize
			backend.VFS.BackendRemovedSize = vnode.Backend.SummaryStats.RecordsRemovedSize
			backend.VFS.BackendUsedSize = vnode.Backend.SummaryStats.BaseSize

			for cname, cstat := range vnode.Commands {
				backend.Commands[cname] = &CStat {
					RequestsSuccess:	cstat.RequestsSuccess(),
					RequestsFailures:	cstat.RequestsFailures(),
					Bytes:			cstat.Bytes(),
				}
			}
		}
	}

	return
}

func (stat *DnetStat) Finalize() {
	for _, sg := range stat.Group {
		sg.Finalize()
	}
}

func (stat *DnetStat) StatData() (reply map[string]interface{}) {
	reply = make(map[string]interface{})
	for group, sg := range stat.Group {
		reply[fmt.Sprintf("%d", group)] = sg.StatGroupData()
	}
	return reply
}
