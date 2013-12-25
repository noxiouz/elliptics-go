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

#ifndef SESSION_H
#define SESSION_H

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
struct go_read_result 
{
	char *file;
	size_t size;
	struct dnet_io_attr *io_attribute;
};

//lookup_result_entry
struct go_lookup_result
{
	struct dnet_file_info *info;
	struct dnet_addr *addr;
	const char *path;
};


//index_entry
struct c_index_entry
{
	const char *data;
	size_t size;
};

//find_indexes_result_entry
struct go_find_result
{
	const struct dnet_raw_id *id;
	size_t entries_count;
	struct c_index_entry *entries;
};

typedef void(*gocallback)(void*, void*);

ell_session* new_elliptics_session(ell_node* node);

void session_set_groups(ell_session *session, int32_t* groups, int count);
void session_set_namespace(ell_session *session, const char *name, int nsize);

void session_lookup(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key);

void session_read_data(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key);
void session_write_data(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key, char *data, size_t size);

void session_remove(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key);

void session_find_all_indexes(ell_session *session, void *on_chunk_context, void *final_context, char *indexes[], size_t nsize);
void session_find_any_indexes(ell_session *session, void *on_chunk_context, void *final_context, char *indexes[], size_t nsize);


#ifdef __cplusplus 
}
#endif

#endif
