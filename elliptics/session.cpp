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

#include "session.h"
#include <errno.h>

using namespace ioremap;

extern "C" {

//This header is generated by cgo in compile time
#include "_cgo_export.h"

struct go_data_pointer new_data_pointer(char *data, int size) {
	return go_data_pointer{data, size};
}

void on_finish(void *context, const elliptics::error_info &error) {
	go_final_callback(error.code(), context);
}

ell_session* new_elliptics_session(ell_node* node)
{
	ell_session *session = new elliptics::session(*node);
	session->set_exceptions_policy(elliptics::session::no_exceptions);
	return session;
}

void session_set_groups(ell_session *session, int32_t *groups, int count)
{
	std::vector<int> g(groups, groups + count);
	session->set_groups(g);
}

void session_set_namespace(ell_session *session, const char *name, int nsize)
{
	session->set_namespace(name, nsize);
}

void session_set_timeout(ell_session *session, int timeout)
{
	session->set_timeout(timeout);
}

/*
	Read
*/
void on_read(void *context, const elliptics::read_result_entry &result)
{
	elliptics::data_pointer data(result.file());
	go_read_result to_go{
		(char *)data.data(),
		data.size(),
		result.io_attribute()};
	go_read_callback(&to_go, context);
}

void session_read_data(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key)
{
	using namespace std::placeholders;
	session->read_data(*key, 0, 0).connect(std::bind(&on_read, on_chunk_context, _1), std::bind(&on_finish, final_context, _1));
}

/*
	Write and Lookup
*/
void on_lookup(void *context, const elliptics::lookup_result_entry &result)
{
	go_lookup_result to_go{
		result.file_info(),
		result.storage_address(),
		result.file_path()
	};
	go_lookup_callback(&to_go, context);
}

void session_write_data(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key, char *data, size_t size)
{
	using namespace std::placeholders;

	std::string tmp(data, size);
	session->write_data(*key, tmp, 0).connect(
			std::bind(&on_lookup, on_chunk_context, _1),
			std::bind(&on_finish, final_context, _1));
}

void session_lookup(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key)
{
	using namespace std::placeholders;
	session->lookup(*key).connect(std::bind(&on_lookup, on_chunk_context, _1), std::bind(&on_finish, final_context, _1));
}

/*
	Remove
*/
// Not implemented. Don't know about anything usefull informaitopn from result.
void on_remove(void *context, const elliptics::remove_result_entry &result)
{
	(void) result; 
	(void) context;
}

void session_remove(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key)
{
	using namespace std::placeholders;
	session->remove(*key).connect(std::bind(&on_remove, on_chunk_context, _1), std::bind(&on_finish, final_context, _1));
}

/*
	Find
*/
void on_find(void *context, const elliptics::find_indexes_result_entry &result) {
	std::vector<c_index_entry> c_index_entries;
	for(size_t i = 0; i< result.indexes.size(); i++)
	{
		c_index_entries.push_back(
			c_index_entry{
				(const char *)result.indexes[i].data.data(),
				result.indexes[i].data.size()}
			);
	}
	go_find_result to_go{
		&result.id,
		c_index_entries.size(),
		c_index_entries.data()
	};

	go_find_callback(&to_go, context);
}

void session_find_all_indexes(ell_session *session, void *on_chunk_context, void *final_context, char *indexes[], size_t nsize)
{
	using namespace std::placeholders;
	std::vector<std::string> index_names;
	index_names.reserve(nsize);
	for(size_t i = 0; i < nsize; i++)
	{
		index_names.push_back(indexes[i]);
	}
	session->find_all_indexes(index_names).connect(
			std::bind(&on_find, on_chunk_context, _1),
			std::bind(&on_finish, final_context, _1));
}

void session_find_any_indexes(ell_session *session, void *on_chunk_context, void *final_context, char *indexes[], size_t nsize)
{
	using namespace std::placeholders;
	std::vector<std::string> index_names;
	index_names.reserve(nsize);
	for(size_t i = 0; i < nsize; i++)
	{
		index_names.push_back(indexes[i]);
	}
	session->find_any_indexes(index_names).connect(
			std::bind(&on_find, on_chunk_context, _1),
			std::bind(&on_finish, final_context, _1));
}

/*
	Indexes
*/
// Not implemented. Don't know about anything usefull informaitopn from result.
void on_set_indexes(void *context, const elliptics::callback_result_entry &result) {
	(void) context;
	(void) result;
}

void on_list_indexes(void *context, const elliptics::index_entry &result) {
	c_index_entry to_go{
		(const char *)result.data.data(),
		result.data.size()};
	go_index_entry_callback(&to_go, context);
}

void session_list_indexes(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key)
{
	using namespace std::placeholders;
	session->list_indexes(*key).connect(std::bind(&on_list_indexes, on_chunk_context, _1),
										std::bind(&on_finish, final_context, _1));
}

void session_set_indexes(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key,
						 char *indexes[], struct go_data_pointer *data, size_t count) 
{
	/* Move to util function */
	using namespace std::placeholders;
	std::vector<std::string> index_names;
	index_names.reserve(count);
	std::vector<elliptics::data_pointer> index_datas;
	index_datas.reserve(count);
	for(size_t i = 0; i < count; i++)
	{
		index_names.push_back(indexes[i]);
		index_datas.emplace_back(elliptics::data_pointer::copy(data[i].data, data[i].size));
	}
	session->set_indexes(*key, index_names, index_datas).connect(
			std::bind(&on_set_indexes, on_chunk_context, _1),
			std::bind(&on_finish, final_context, _1));
}

void session_update_indexes(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key,
						 char *indexes[], struct go_data_pointer *data, size_t count) 
{
	/* Move to util function */
	using namespace std::placeholders;
	std::vector<std::string> index_names;
	index_names.reserve(count);
	std::vector<elliptics::data_pointer> index_datas;
	index_datas.reserve(count);
	for(size_t i = 0; i < count; i++)
	{
		index_names.push_back(indexes[i]);
		index_datas.emplace_back(elliptics::data_pointer::copy(data[i].data, data[i].size));
	}
	session->update_indexes(*key, index_names, index_datas).connect(
			std::bind(&on_set_indexes, on_chunk_context, _1),
			std::bind(&on_finish, final_context, _1));
}

void session_remove_indexes(ell_session *session, void *on_chunk_context, void *final_context, ell_key *key, char *indexes[], size_t nsize)
{
	using namespace std::placeholders;
	std::vector<std::string> index_names;
	index_names.reserve(nsize);
	for(size_t i = 0; i < nsize; i++)
	{
		index_names.push_back(indexes[i]);
	}
	session->remove_indexes(*key, index_names).connect(
			std::bind(&on_set_indexes, on_chunk_context, _1),
			std::bind(&on_finish, final_context, _1));
}

} // extern "C"
