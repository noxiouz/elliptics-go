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
)


type Time struct {
	Sec		uint64		`json:"tv_sec"`
	USec		uint64		`json:"tv_usec"`
}
type Status struct {
	State		int		`json:"state"`
	DefragState	int		`json:"defrag_state"`
	LastStart	Time		`json:"last_start"`
	LastStartErr	int		`json:"last_start_err"`
	RO		bool		`json:"read_only"`
	Delay		uint32		`json:"delay"`
}

type Config struct {
	Group		uint32		`json:"group"`
	Data		string		`json:"data"`
	Sync		int		`json:"sync"`
	BlobFlags	uint64		`json:"blob_flags"`
	BlobSize	uint64		`json:"blob_size"`
	BlobSizeLimit	uint64		`json:"blob_size_limit"`
	MaxRecords	uint64		`json:"records_in_blob"`
	DefragPercentage	uint64	`json:"defrag_percentage"`
	DefragTimeout	uint64		`json:"defrag_timeout"`
	DefragTime	uint64		`json:"defrag_time"`
	DefragSplay	uint64		`json:"defrag_splay"`
}
type GlobalStats struct {
	DataSortStartTime		uint64		`json:"datasort_start_time"`
	DataSortCompletionTime		uint64		`json:"datasort_completion_time"`
	DataSortCompletionStatus	int		`json:"datasort_completion_status"`
}
type BlobStats struct {
	RecordsTotal		uint64		`json:"records_total"`
	RecordsRemoved		uint64		`json:"records_removed"`
	RecordsRemovedSize	uint64		`json:"records_removed_size"`
	RecordsCorrupted	uint64		`json:"records_corrupted"`
	BaseSize		uint64		`json:"base_size"`
	WantDefrag		int		`json:"want_defrag"`
	IsSorted		int		`json:"is_sorted"`
}
type VStat struct {
	BSize			uint64		`json:"bsize"`
	FrSize			uint64		`json:"frsize"`
	Blocks			uint64		`json:"blocks"`
	BFree			uint64		`json:"bfree"`
	BAvail			uint64		`json:"bavail"`
}
type DStatRaw struct {
	ReadIOs			uint64		`json:"read_ios"`
	ReadMerges		uint64		`json:"read_merges"`
	ReadSectors		uint64		`json:"read_sectors"`
	ReadTicks		uint64		`json:"read_ticks"`
	WriteIOs		uint64		`json:"write_ios"`
	WriteMerges		uint64		`json:"write_merges"`
	WriteSectors		uint64		`json:"write_sectors"`
	WriteTicks		uint64		`json:"write_ticks"`
	InFlight		uint64		`json:"in_flight"`
	IOTicks			uint64		`json:"io_ticks"`
	TimeInQueue		uint64		`json:"time_in_queue"`
}
type BackendError struct {
	Code			int		`json:"code"`
}
type Backend struct {
	Config	Config				`json:"config"`
	GlobalStats GlobalStats			`json:"global_stats"`
	SummaryStats BlobStats			`json:"summary_stats"`
	BaseStats map[string]BlobStats		`json:"base_stats"`
	VFS VStat				`json:"vfs"`
	DStat DStatRaw				`json:"dstat"`
	Error BackendError			`json:"error"`
}

type CommandStat struct {
	Success		uint64			`json:"successes"`
	Failures	uint64			`json:"failures"`
	Size		uint64			`json:"size"`
	Time		uint64			`json:"time"`
}
func (c *CommandStat) RequestsSuccess() uint64 {
	return c.Success
}
func (c *CommandStat) RequestsFailures() uint64 {
	return c.Failures
}
func (c *CommandStat) Bytes() uint64 {
	return c.Size
}

type LayerStat struct {
	Outside		CommandStat		`json:"outside"`
	Internal	CommandStat		`json:"internal"`
}
func (l *LayerStat) RequestsSuccess() uint64 {
	return l.Outside.RequestsSuccess() + l.Internal.RequestsSuccess()
}
func (l *LayerStat) RequestsFailures() uint64 {
	return l.Outside.RequestsFailures() + l.Internal.RequestsFailures()
}
func (l *LayerStat) Bytes() uint64 {
	return l.Outside.Bytes() + l.Internal.Bytes()
}

type PacketOnlyCommandStat struct {
	Success		uint64			`json:"successes"`
	Failures	uint64			`json:"failures"`
}
type DstStat struct {
	Storage		PacketOnlyCommandStat	`json:"storage"`
	Proxy		PacketOnlyCommandStat	`json:"proxy"`
}
type Command struct {
	Cache		LayerStat		`json:"cache"`
	Disk		LayerStat		`json:"disk"`
	Total		DstStat			`json:"total"`
}
func (c *Command) RequestsSuccess() uint64 {
	return c.Cache.RequestsSuccess() + c.Disk.RequestsSuccess()
}
func (c *Command) RequestsFailures() uint64 {
	return c.Cache.RequestsFailures() + c.Disk.RequestsFailures()
}
func (c *Command) Bytes() uint64 {
	return c.Cache.Bytes() + c.Disk.Bytes()
}

type VNode struct {
	BackendID	int			`json:"backend_id"`
	Status		Status			`json:"status"`
	Backend		Backend			`json:"backend"`
	Commands	map[string]Command	`json:"commands"`
}

type Response struct {
	Timestamp	Time			`json:"timestamp"`
	MonitorStatus	string			`json:"monitor_status"`
	Backends	map[string]VNode	`json:"backends"`
	Commands	map[string]Command	`json:"commands"`
}

