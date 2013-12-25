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

#include "_cgo_export.h"

void on_finish(void *context, const elliptics::error_info &error) {
	go_final_callback(error.code(), context);
}

void on_write_result(void *context, const elliptics::sync_write_result &result, const elliptics::error_info &error)
{
	if (error) {
		//go_read_callback(NULL, 0, error.code(), context);
	} else {
		std::vector<go_write_result> to_go;

		for (size_t i = 0 ; i < result.size(); i++)
		{
			go_write_result tmp;
			tmp.info = result[i].file_info();
			tmp.addr = result[i].storage_address();
			tmp.path = result[i].file_path();
			to_go.push_back(tmp);
		}

		go_lookup_callback(&to_go[0], result.size(), error.code(), context);
	}
}

void on_remove(void *context, const elliptics::sync_remove_result &result, const elliptics::error_info &error)
{
	(void) result;
	go_remove_callback(error.code(), context);
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
	session->read_data(*key, 0, 0).connect(std::bind(&on_read, on_chunk_context, _1),
										   std::bind(&on_finish, final_context, _1));
}

/*
	Write
*/
void session_write_data(ell_session *session, void *context, ell_key *key, char *data, size_t size)
{
	using namespace std::placeholders;
	session->write_data(*key, elliptics::data_pointer(data, size), 0).connect(std::bind(&on_write_result, context, _1, _2));
}

/*
	Remove
*/
void session_remove(ell_session *session, void *context, ell_key *key)
{
	using namespace std::placeholders;
	session->remove(*key).connect(std::bind(&on_remove, context, _1, _2));
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
	session->find_all_indexes(index_names).connect(std::bind(&on_find, on_chunk_context, _1),
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
	session->find_any_indexes(index_names).connect(std::bind(&on_find, on_chunk_context, _1),
												   std::bind(&on_finish, final_context, _1));
}
} // extern "C"
