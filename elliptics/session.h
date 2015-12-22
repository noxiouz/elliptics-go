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

#ifndef __ELLIPTICS_SESSION_H
#define __ELLIPTICS_SESSION_H

#include "node.h"
#include "key.h"

typedef uint64_t context_t;

#ifdef __cplusplus

#include <iostream>
#include <elliptics/session.hpp>
typedef ioremap::elliptics::session ell_session;

extern "C" {
#else
typedef void ell_session;
#endif

//read_result_entry
struct go_read_result {
	const struct dnet_cmd		*cmd;
	const struct dnet_addr		*addr;
	const struct dnet_io_attr	*io_attribute;
	const char			*file;
	uint64_t			size;
};

//lookup_result_entry
struct go_lookup_result {
	const struct dnet_cmd		*cmd;
	const struct dnet_addr		*addr;
	const struct dnet_file_info	*info;
	const struct dnet_addr		*storage_addr;
	const char			*path;
};

//remove_result_entry
struct go_remove_result {
	const struct dnet_cmd		*cmd;
};

//index_entry
struct c_index_entry {
	const char			*data;
	uint64_t			size;
};

//find_indexes_result_entry
struct go_find_result {
	const struct dnet_raw_id	*id;
	const uint64_t			entries_count;
	struct c_index_entry		*entries;
};

struct go_data_pointer {
	char *data;
	int size;
};

struct go_error {
	int				code;		// elliptics error code, should be negative errno value
	uint64_t			flags;		// dnet_cmd.flags
	const char			*message;
};


struct go_iterator_range {
	uint8_t	*key_begin;
	uint8_t *key_end;
};

struct go_iterator_result {
	// dnet_iterator_response
	struct dnet_iterator_response	*reply;

	// data_pointer
	const char			*reply_data;
	const uint64_t			reply_size;

	// iterator_id
	uint64_t			id;
};


struct go_data_pointer new_data_pointer(char *data, int size);

ell_session *new_elliptics_session(ell_node *node);
void delete_session(ell_session *session);

const char *session_transform(ell_session *session, const char *key);

void session_set_filter_all(ell_session *session);
void session_set_filter_positive(ell_session *session);

void session_set_groups(ell_session *session, uint32_t *groups, int count);
void session_set_namespace(ell_session *session, const char *name, int nsize);

void session_set_timeout(ell_session *session, int timeout);
long session_get_timeout(ell_session *session);

typedef uint64_t cflags_t;
void session_set_cflags(ell_session *session, cflags_t cflags);
cflags_t session_get_cflags(ell_session *session);

typedef uint32_t ioflags_t;
void session_set_ioflags(ell_session *session, ioflags_t ioflags);
ioflags_t session_get_ioflags(ell_session *session);

void session_set_trace_id(ell_session *session, trace_id_t trace_id);
trace_id_t session_get_trace_id(ell_session *session);

// ->lookup() returns only the first group where given key has been found
void session_lookup(ell_session *session, context_t on_chunk_context,
		context_t final_context, ell_key *key);
// ->parallel_lookup() sends multiple lookups in parallel and returns all groups where given key has been found
void session_parallel_lookup(ell_session *session, context_t on_chunk_context,
		context_t final_context, ell_key *key);

void session_read_data_into(ell_session *session, context_t on_chunk_context, context_t buf_context,
		context_t final_context, ell_key *key, uint64_t offset, uint64_t size);
void session_read_data(ell_session *session, context_t on_chunk_context,
		context_t final_context, ell_key *key, uint64_t offset, uint64_t size);
void session_write_data(ell_session *session, context_t on_chunk_context,
		context_t final_context, ell_key *key, uint64_t offset, char *data, uint64_t size);

// prepare/write/commit sequence for large objects
// @offset says on which offset should data go
// @total_size is a size of contiguous disk area to prepare, i.e. disk_size, data_size (or read record size) can be smaller than that
// @commit_size means how many bytes to actually commit as record size, if 0 all written data will be used
//
// Each of these calls can be accomplished with data chunk
void session_write_prepare(ell_session *session, context_t on_chunk_context,
			context_t final_context, ell_key *key,
			uint64_t offset, uint64_t total_size,
			char *data, uint64_t size);
void session_write_plain(ell_session *session, context_t on_chunk_context,
			context_t final_context, ell_key *key,
			uint64_t offset,
			char *data, uint64_t size);
void session_write_commit(ell_session *session, context_t on_chunk_context,
			context_t final_context, ell_key *key,
			uint64_t offset,
			uint64_t commit_size,
			char *data, uint64_t size);

void session_remove(ell_session *session, context_t on_chunk_context,
		context_t final_context, ell_key *key);
void session_bulk_remove(ell_session *session, context_t on_chunk_context, context_t final_context, void *ekeys);

void session_find_all_indexes(ell_session *session, context_t on_chunk_context,
		context_t final_context, char *indexes[], uint64_t nsize);
void session_find_any_indexes(ell_session *session, context_t on_chunk_context,
		context_t final_context, char *indexes[], uint64_t nsize);

void session_set_indexes(ell_session *session, context_t on_chunk_context,
		context_t final_context, ell_key *key, char *indexes[],
		struct go_data_pointer *data, uint64_t count);

void session_update_indexes(ell_session *session, context_t on_chunk_context,
		context_t final_context, ell_key *key,
		char *indexes[], struct go_data_pointer *data, uint64_t count);

void session_remove_indexes(ell_session *session, context_t on_chunk_context,
		context_t final_context, ell_key *key,
		char *indexes[], uint64_t nsize);

void session_list_indexes(ell_session *sesion, context_t on_chunk_context,
		context_t final_context, ell_key *key);

struct go_backends_status {
	struct dnet_backend_status_list		*list;
	struct go_error				error;
};

void session_backends_status(ell_session *session, const struct dnet_addr *addr, context_t context);
void session_backend_start_defrag(ell_session *session, const struct dnet_addr *addr, uint32_t backend_id, context_t context);
void session_backend_enable(ell_session *session, const struct dnet_addr *addr, uint32_t backend_id, context_t context);
void session_backend_disable(ell_session *session, const struct dnet_addr *addr, uint32_t backend_id, context_t context);
void session_backend_make_writable(ell_session *session, const struct dnet_addr *addr, uint32_t backend_id, context_t context);
void session_backend_make_readonly(ell_session *session, const struct dnet_addr *addr, uint32_t backend_id, context_t context);
void session_backend_set_delay(ell_session *session, const struct dnet_addr *addr, uint32_t backend_id, uint32_t delay, context_t context);

int session_lookup_addr(ell_session *session, const char *key, int len, int group_id, struct dnet_addr *addr, int *backend_id);

static inline struct dnet_addr *dnet_addr_alloc()
{
	struct dnet_addr *addr = (struct dnet_addr *)malloc(sizeof(struct dnet_addr));
	memset(addr, 0, sizeof(struct dnet_addr));
	return addr;
}

static inline void dnet_addr_free(struct dnet_addr *addr)
{
	free(addr);
}

void session_start_iterator(ell_session *session, context_t on_chunk_context, context_t final_context,
			const struct go_iterator_range* ranges, size_t range_count,
			const ell_key *key,
			uint64_t type,
			uint64_t flags,
			struct dnet_time time_begin,
			struct dnet_time time_end);

void session_start_copy_iterator(ell_session *session, context_t on_chunk_context, context_t final_context,
			const struct go_iterator_range* ranges, size_t range_count,
			uint32_t *groups, size_t groups_count,
			const ell_key *key,
			uint64_t flags,
			struct dnet_time time_begin,
			struct dnet_time time_end);


void session_pause_iterator(ell_session *session, context_t on_chunk_context, context_t final_context,
			ell_key *key,
			uint64_t iterator_id);

void session_continue_iterator(ell_session *session, context_t on_chunk_context, context_t final_context,
			ell_key *key,
			uint64_t iterator_id);

void session_cancel_iterator(ell_session *session, context_t on_chunk_context, context_t final_context,
			ell_key *key,
			uint64_t iterator_id);

void session_server_send(ell_session *session, context_t on_chunk_context, context_t final_context,
			void *ekeys,
			uint64_t flags,
			uint32_t *groups, size_t groups_count);

#ifdef __cplusplus
}
#endif

#endif /* __ELLIPTICS_SESSION_H */
