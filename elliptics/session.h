/*
 * 2013+ Copyright (c) Anton Tyurin <noxiouz@yandex.ru>
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

#ifndef __ELLIPTICS_SESSION_H
#define __ELLIPTICS_SESSION_H

#include "node.h"
#include "key.h"

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
	char				*file;
	size_t				size;
};

//lookup_result_entry
struct go_lookup_result {
	const struct dnet_cmd		*cmd;
	const struct dnet_addr		*addr;
	const struct dnet_file_info	*info;
	const struct dnet_addr		*storage_addr;
	const char			*path;
};

//index_entry
struct c_index_entry {
	const char		*data;
	size_t			size;
};

//find_indexes_result_entry
struct go_find_result {
	const struct dnet_raw_id	*id;
	const size_t			entries_count;
	struct c_index_entry		*entries;
};

struct go_data_pointer {
	char *data;
	int size;
};

struct go_error {
	int		code;		// elliptics error code, should be negative errno value
	int		flags;		// this will mainly say whether it is client or server error in 2.26
	const char	*message;
};

struct go_data_pointer new_data_pointer(char *data, int size);

ell_session *new_elliptics_session(ell_node *node);

void session_set_groups(ell_session *session, int32_t *groups, int count);
void session_set_namespace(ell_session *session, const char *name, int nsize);
void session_set_timeout (ell_session *session, int timeout);

void session_lookup(ell_session *session, void *on_chunk_context,
		void *final_context, ell_key *key);

void session_read_data(ell_session *session, void *on_chunk_context,
		void *final_context, ell_key *key);
void session_write_data(ell_session *session, void *on_chunk_context,
		void *final_context, ell_key *key, char *data, size_t size);

void session_remove(ell_session *session, void *on_chunk_context,
		void *final_context, ell_key *key);

void session_find_all_indexes(ell_session *session, void *on_chunk_context,
		void *final_context, char *indexes[], size_t nsize);
void session_find_any_indexes(ell_session *session, void *on_chunk_context,
		void *final_context, char *indexes[], size_t nsize);

void session_set_indexes(ell_session *session, void *on_chunk_context,
		void *final_context, ell_key *key, char *indexes[],
		struct go_data_pointer *data, size_t count);

void session_update_indexes(ell_session *session, void *on_chunk_context,
		void *final_context, ell_key *key,
		char *indexes[], struct go_data_pointer *data, size_t count);

void session_remove_indexes(ell_session *session, void *on_chunk_context,
		void *final_context, ell_key *key,
		char *indexes[], size_t nsize);

void session_list_indexes(ell_session *sesion, void *on_chunk_context,
		void *final_context, ell_key *key);

#ifdef __cplusplus
}
#endif

#endif /* __ELLIPTICS_SESSION_H */
