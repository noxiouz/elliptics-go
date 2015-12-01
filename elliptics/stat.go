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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"
)

/*
#include "stat.h"
*/
import "C"

const (
	StatCategoryCache    int64  = 1 << 0
	StatCategoryIO       int64  = 1 << 1
	StatCategoryCommands int64  = 1 << 2
	StatCategoryBackend  int64  = 1 << 4
	StatCategoryProcFS   int64  = 1 << 6
	StatSectorSize       uint64 = 512

	BackendStateDisabled     int32 = 0
	BackendStateEnabled      int32 = 1
	BackendStateActivating   int32 = 2
	BackendStateDeactivating int32 = 3

	DefragStateNotStarted int32 = 0
	DefragStateInProgress int32 = 1
)

var (
	backend_state = map[int32]string{
		BackendStateDisabled:     "disabled",
		BackendStateEnabled:      "enabled",
		BackendStateActivating:   "activating",
		BackendStateDeactivating: "deactivating",
	}
	defrag_state = map[int32]string{
		DefragStateNotStarted: "not-started",
		DefragStateInProgress: "in-progress",
	}
)

type RawAddr struct {
	Addr   [32]byte
	Len    int
	Family uint16
}

func (a *RawAddr) DnetAddr() *DnetAddr {
	return &DnetAddr{
		Addr:   a.Addr[:a.Len],
		Family: a.Family,
	}
}
func (a *RawAddr) String() string {
	tmp := a.DnetAddr()
	return tmp.String()
}

type VFS struct {
	// space in bytes for given backend
	Total, Avail uint64

	// logical size limitation for backends which support it
	// blob backend may set this (configuration must allow blob size checks, bit-4 must be zero)
	// for all others this field equals to @VFS.Total
	TotalSizeLimit uint64

	BackendRemovedSize uint64
	BackendUsedSize    uint64

	RecordsTotal     uint64
	RecordsRemoved   uint64
	RecordsCorrupted uint64
}

type CStat struct {
	RequestsSuccess  uint64
	RequestsFailures uint64
	Bytes            uint64

	RPSFailures float64
	RPSSuccess  float64
	BPS         float64
}

type DStat struct {
	WSectors uint64
	RSectors uint64
	IOTicks  uint64

	WBS  float64
	RBS  float64
	Util float64
}

type PID struct {
	sync.RWMutex

	Error         float64
	IntegralError float64
	ErrorTime     time.Time
	Pain          float64
}

func NewPIDController() PID {
	return PID{
		ErrorTime: time.Now(),
	}
}

const (
	PIDKe float64 = 1.0

	// Integral part has to be zero in this case
	// since there is no continuous 'force' to check/change in our case
	// we can not infinitely increase integral part in attempt to compensate
	// for difference of the error from zero (or no-matter-what like 100 MB/s)
	// integral part has to compensate speed of wind when we are trying to achieve
	// desired velocity, but in our case there is no engine controller which
	// can output continuous power to driver the vehicle, instead we have to
	// determine which of the backends is currently the fastest
	PIDKi float64 = 0
	PIDKd float64 = 0.3
)

type AddressBackend struct {
	Addr    RawAddr
	Backend int32
}

func (ab *AddressBackend) String() string {
	return fmt.Sprintf("ab: %s/%d", ab.Addr.String(), ab.Backend)
}

func NewAddressBackend(addr *DnetAddr, backend int32) AddressBackend {
	raw := RawAddr {
		Len:    len(addr.Addr),
		Family: addr.Family,
	}
	if len(addr.Addr) > len(raw.Addr) {
		log.Fatalf("can not create new address+backend: addr: %s, backend: %d, addr.len: %d, raw.len: %d\n\n",
			addr.String(), backend, len(addr.Addr), len(raw.Addr))
	}

	copy(raw.Addr[:], addr.Addr)

	return AddressBackend{
		Addr:    raw,
		Backend: backend,
	}
}

func (entry *StatEntry) AddressBackend() AddressBackend {
	return NewAddressBackend(&entry.addr, entry.cmd.Backend)
}


type StatBackend struct {
	// Address+Backend for given stats, do not put it into json
	// since @RawAddr is an array of bytes and it can not be parsed by human
	Ab AddressBackend `json:"-"`

	Error BackendError `json:"error"`

	// All range starts (IDs) for given node (server address + backend)
	ID []DnetRawID `json:"-"`

	// @sum hosts total sum of uint64 ids created frm DnetRawID
	// after it is devided by math.MaxUint64 it becomes @Percentage
	sum uint64

	// percentage of the whole IDs ring currently occupied by given ids
	// and thus node (address + backend)
	Percentage float64

	// defragmentation status: 0 - not started, 1 - in progress
	DefragState            int32
	DefragStateStr         string
	DefragStartTime        time.Time
	DefragCompletionTime   time.Time
	DefragCompletionStatus int32

	// backend is in read-only mode
	RO bool

	// backend has delay of @Delay ms for every operation
	Delay uint32

	// VFS statistics: available, used and total space
	VFS VFS

	// PID-controller used for data writing
	PID PID

	// per-command size/number counters
	// difference between the two divided by the time difference equals to RPS/BPS
	Commands map[string]*CStat
}

func NewStatBackend(ab AddressBackend) *StatBackend {
	return &StatBackend {
		Ab:		ab,
		ID:		make([]DnetRawID, 0),
		Percentage:	0,
		sum:		0,
		Commands:	make(map[string]*CStat),
		PID:		NewPIDController(),
	}
}

func (backend *StatBackend) PIDPain() float64 {
	p := &backend.PID

	p.RLock()
	defer p.RUnlock()

	return p.Pain
}
func (backend *StatBackend) PIDUpdate(e float64) {
	p := &backend.PID
	p.Lock()
	defer p.Unlock()

	delta_T := time.Since(p.ErrorTime).Seconds()

	if delta_T == 0 {
		return
	}

	integral_new := e*delta_T + p.IntegralError
	diff := (e - p.Error) / delta_T

	u := e*PIDKe + integral_new*PIDKi + diff*PIDKd
	if u <= 0 {
		// negative 'force' means we have to decrease/to hold down current attempts
		// to achieve desired performance, but we do not have engine to tune its power,
		// instead we have to select the fastest backend
		// thus, let's just reduce the pain of the current backend when its performance
		// is higher than desired one
		u = p.Pain/2.0 + 1.0
	}

	p.IntegralError = integral_new
	p.Error = e
	p.ErrorTime = time.Now()
	p.Pain = u
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

func (sg *StatGroup) FindStatBackendKey(s *Session, key string, group_id uint32) (*StatBackend, error) {
	addr, backend_id, err := s.LookupBackend(key, group_id)
	if err != nil {
		return nil, &DnetError{
			Code:    -2, // -ENOENT
			Flags:   0,
			Message: fmt.Sprintf("could not find backend for key: %s, group: %d: %v", key, group_id, err),
		}
	}

	return sg.FindStatBackend(addr, backend_id)
}

func (sg *StatGroup) FindStatBackend(addr *DnetAddr, backend_id int32) (*StatBackend, error) {
	ab := NewAddressBackend(addr, backend_id)

	st, ok := sg.Ab[ab]
	if !ok {
		return nil, &DnetError{
			Code:  -2, // -ENOENT
			Flags: 0,
			Message: fmt.Sprintf("could not find statistics for addr: %s, backend: %d",
				addr.String(), backend_id),
		}
	}

	return st, nil
}

type StatBackendData struct {
	Address string
	Backend int32
	Stat    *StatBackend
}

type StatGroupData struct {
	Backends []*StatBackendData
}

func (sg *StatGroup) StatGroupData() (reply *StatGroupData) {

	reply = &StatGroupData{
		Backends: make([]*StatBackendData, 0, len(sg.Ab)),
	}
	for ab, backend := range sg.Ab {
		tmp := &StatBackendData{
			Address: ab.Addr.String(),
			Backend: ab.Backend,
			Stat:    backend,
		}

		reply.Backends = append(reply.Backends, tmp)
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
	Time  time.Time
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

//export go_stat_callback
func go_stat_callback(result *C.struct_go_stat_result, key uint64) {
	context, err := Pool.Get(key)
	if err != nil {
		panic("Unable to find session numbder")
	}
	callback := context.(func(*StatEntry))

	res := &StatEntry{
		cmd:  NewDnetCmd(result.cmd),
		addr: NewDnetAddr(result.addr),
		err:  nil,
	}

	if result.stat_data != nil && result.stat_size != 0 {
		res.stat = C.GoBytes(unsafe.Pointer(result.stat_data), C.int(result.stat_size))
	} else {
		res.stat = make([]byte, 0)
	}

	callback(res)
}

func (s *Session) DnetStat() *DnetStat {
	response := make(chan *StatEntry, 10)

	onResultContext := NextContext()
	onFinishContext := NextContext()

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
		Pool.Delete(onResultContext)
		Pool.Delete(onFinishContext)
	}

	Pool.Store(onResultContext, onResult)
	Pool.Store(onFinishContext, onFinish)

	categories := StatCategoryBackend | StatCategoryProcFS | StatCategoryCommands

	C.session_get_stats(s.session,
		C.context_t(onResultContext), C.context_t(onFinishContext),
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
	if prev == nil || prev == stat {
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

			sb.PID = psb.PID

			for cmd, cstat := range sb.Commands {
				pcstat, ok := psb.Commands[cmd]
				if !ok {
					continue
				}

				cstat.RPSSuccess = float64(cstat.RequestsSuccess-pcstat.RequestsSuccess) / duration
				cstat.RPSFailures = float64(cstat.RequestsFailures-pcstat.RequestsFailures) / duration
				cstat.BPS = float64(cstat.Bytes-pcstat.Bytes) / duration
			}
		}
	}
}

func (stat *DnetStat) FindCreateBackend(group uint32, addr *DnetAddr, backend_id int32) *StatBackend {
	sg, ok := stat.Group[group]
	if !ok {
		nsg := NewStatGroup()
		sg = &nsg
		stat.Group[group] = sg
	}

	ab := NewAddressBackend(addr, backend_id)

	backend, ok := sg.Ab[ab]
	if !ok {
		backend = NewStatBackend(ab)
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

	var r Response

	err := json.Unmarshal(entry.stat, &r)
	if err != nil {
		log.Printf("%s: could not parse stat entry '%s' reply: %v\n", entry.addr.String(), string(entry.stat), err)
	}

	stat.Time = time.Unix(int64(r.Timestamp.Sec), int64(r.Timestamp.USec*1000))

	if r.MonitorStatus != "enabled" {
		log.Printf("%s: monitoring doesn't work: %v\n", entry.addr.String(), r.MonitorStatus)
		return
	}

	good_backends := 0
	for _, vnode := range r.Backends {
		if vnode.Status.State != BackendStateEnabled {
			log.Printf("stat: addr: %s, backend: %d, group: %d: DISABLED\n",
				entry.addr.String(), int32(vnode.BackendID), vnode.Backend.Config.Group, vnode.Backend.Error.Code)
			// do not update backend statistics
			continue
		}

		backend := stat.FindCreateBackend(vnode.Backend.Config.Group, &entry.addr, int32(vnode.BackendID))

		if vnode.Backend.Error.Code != 0 {
			log.Printf("stat: addr: %s, backend: %d, group: %d: ERROR: %d\n",
				entry.addr.String(), int32(vnode.BackendID), vnode.Backend.Config.Group, vnode.Backend.Error.Code)
			backend.Error = vnode.Backend.Error
		}

		backend.VFS.Total = vnode.Backend.VFS.FrSize * vnode.Backend.VFS.Blocks
		backend.VFS.Avail = vnode.Backend.VFS.BFree * vnode.Backend.VFS.BSize
		backend.VFS.BackendRemovedSize = vnode.Backend.SummaryStats.RecordsRemovedSize
		backend.VFS.BackendUsedSize = vnode.Backend.SummaryStats.BaseSize
		backend.VFS.RecordsTotal = vnode.Backend.SummaryStats.RecordsTotal
		backend.VFS.RecordsRemoved = vnode.Backend.SummaryStats.RecordsRemoved
		backend.VFS.RecordsCorrupted = vnode.Backend.SummaryStats.RecordsCorrupted

		backend.VFS.TotalSizeLimit = backend.VFS.Total
		// check blob flags, if bit 4 is set, blob will not perform size checks at all
		if (vnode.Backend.Config.BlobSizeLimit != 0) && (vnode.Backend.Config.BlobFlags&(1<<4) == 0) {
			backend.VFS.TotalSizeLimit = vnode.Backend.Config.BlobSizeLimit
		}

		//log.Printf("stat: addr: %s, backend: %d, group: %d, used: %d, limit: %d\n",
		//	entry.addr.String(), int32(vnode.BackendID), vnode.Backend.Config.Group,
		//	backend.VFS.BackendUsedSize, backend.VFS.TotalSizeLimit)

		backend.DefragStartTime = time.Unix(int64(vnode.Backend.GlobalStats.DataSortStartTime), 0)
		backend.DefragCompletionTime = time.Unix(int64(vnode.Backend.GlobalStats.DataSortCompletionTime), 0)
		backend.DefragCompletionStatus = vnode.Backend.GlobalStats.DataSortCompletionStatus
		backend.DefragState = vnode.Status.DefragState
		backend.DefragStateStr = defrag_state[vnode.Status.DefragState]
		backend.RO = vnode.Status.RO
		backend.Delay = vnode.Status.Delay

		for cname, cstat := range vnode.Commands {
			backend.Commands[cname] = &CStat{
				RequestsSuccess:  cstat.RequestsSuccess(),
				RequestsFailures: cstat.RequestsFailures(),
				Bytes:            cstat.Bytes(),
			}
		}

		good_backends++
	}

	log.Printf("stat: addr: %s, good-backends: %d/%d\n",
		entry.addr.String(), good_backends, len(r.Backends))

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
