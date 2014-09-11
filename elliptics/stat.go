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
	"log"
	"math"
	"sort"
	"strings"
	"unsafe"
)

/*
#include "stat.h"
*/
import "C"

const (
	StatCategoryCache    int64 = 1 << 0
	StatCategoryIO       int64 = 1 << 0
	StatCategoryCommands int64 = 1 << 2
	StatCategoryBackend  int64 = 1 << 4
	StatCategoryProcFS   int64 = 1 << 6
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

type StatBackend struct {
	// All range starts (IDs) for given node (server address + backend)
	ID []DnetRawID

	// @sum hosts total sum of uint64 ids created frm DnetRawID
	// after it is devided by math.MaxUint64 it becomes @Percentage
	sum uint64

	// percentage of the whole IDs ring currently occupied by given ids
	// and thus node (address + backend)
	Percentage float64

	VFS VFS
}

func NewStatBackend() *StatBackend {
	return &StatBackend{
		ID:         make([]DnetRawID, 0),
		Percentage: 0,
		sum:        0,
	}
}

func (backend *StatBackend) StatBackendData() (reply map[string]interface{}) {
	reply = make(map[string]interface{})

	reply["percentage"] = backend.Percentage
	reply["vfs"] = backend.VFS
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
	Group map[uint32]*StatGroup
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

	categories := StatCategoryBackend | StatCategoryProcFS

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
	)

	type Time struct {
		Sec  uint64 `json:"tv_sec"`
		USec uint64 `json:"tv_usec"`
	}
	type Status struct {
		State        int  `json:"state"`
		DefragState  int  `json:"defrag_state"`
		LastStart    Time `json:"last_start"`
		LastStartErr int  `json:"last_start_err"`
		RO           int  `json:"read_only"`
	}

	type Config struct {
		Group uint32 `json:"group"`
		Data  string `json:"data"`
	}
	type GlobalStats struct {
		DataSortStartTime        uint64 `json:"datasort_start_time"`
		DataSortCompletionTime   uint64 `json:"datasort_completion_time"`
		DataSortCompletionStatus int    `json:"datasort_completion_status"`
	}
	type BlobStats struct {
		RecordsTotal       uint64 `json:"records_total"`
		RecordsRemoved     uint64 `json:"records_removed"`
		RecordsRemovedSize uint64 `json:"records_removed_size"`
		RecordsCorrupted   uint64 `json:"records_corrupted"`
		BaseSize           uint64 `json:"base_size"`
		WantDefrag         int    `json:"want_defrag"`
		IsSorted           int    `json:"is_sorted"`
	}
	type VFS struct {
		BSize  uint64 `json:"bsize"`
		FrSize uint64 `json:"frsize"`
		Blocks uint64 `json:"blocks"`
		BFree  uint64 `json:"bfree"`
		BAvail uint64 `json:"bavail"`
	}
	type DStat struct {
		ReadIOs      uint64 `json:"read_ios"`
		ReadMerges   uint64 `json:"read_merges"`
		ReadSectors  uint64 `json:"read_sectors"`
		ReadTicks    uint64 `json:"read_ticks"`
		WriteIOs     uint64 `json:"write_ios"`
		WriteMerges  uint64 `json:"write_merges"`
		WriteSectors uint64 `json:"write_sectors"`
		WriteTicks   uint64 `json:"write_ticks"`
		InFlight     uint64 `json:"in_flight"`
		IOTicks      uint64 `json:"io_ticks"`
		TimeInQueue  uint64 `json:"time_in_queue"`
	}
	type Backend struct {
		Config       Config               `json:"config"`
		GlobalStats  GlobalStats          `json:"global_stats"`
		SummaryStats BlobStats            `json:"summary_stats"`
		BaseStats    map[string]BlobStats `json:"base_stats"`
		VFS          VFS                  `json:"vfs"`
		DStat        DStat                `json:"dstat"`
	}
	type VNode struct {
		BackendID int     `json:"backend_id"`
		Status    Status  `json:"status"`
		Backend   Backend `json:"backend"`
	}
	type Response struct {
		MonitorStatus string           `json:"monitor_status"`
		Backends      map[string]VNode `json:"backends"`
	}

	var r Response

	err := json.Unmarshal(entry.stat, &r)
	if err != nil {
		log.Printf("%s: could not parse stat entry reply: %v\n", entry.addr.String(), err)
	}

	if r.MonitorStatus != "enabled" {
		log.Printf("%s: monitoring doesn't work: %v\n", entry.addr.String(), r.MonitorStatus)
		return
	}

	for _, vnode := range r.Backends {
		status, ok := backend_state[vnode.Status.State]
		if !ok {
			status = fmt.Sprintf("invalid backend status '%d'", vnode.Status)
		}
		log.Printf("%s: backend: %d: status: %s", entry.addr.String(), vnode.BackendID, status)

		if vnode.Status.State == BackendStateEnabled {
			backend := stat.FindBackend(vnode.Backend.Config.Group, &entry.addr, int32(vnode.BackendID))
			backend.VFS.Total = vnode.Backend.VFS.FrSize * vnode.Backend.VFS.Blocks
			backend.VFS.Avail = vnode.Backend.VFS.BFree * vnode.Backend.VFS.Blocks
			backend.VFS.BackendRemovedSize = vnode.Backend.SummaryStats.RecordsRemovedSize
			backend.VFS.BackendUsedSize = vnode.Backend.SummaryStats.BaseSize
		}
	}

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

func (stat *DnetStat) StatData() (reply map[string]interface{}) {
	reply = make(map[string]interface{})
	for group, sg := range stat.Group {
		reply[fmt.Sprintf("%d", group)] = sg.StatGroupData()
	}
	return reply
}
