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


void on_stat_result(void *context, const elliptics::sync_stat_result &result)
{
	(void) result;
	(void) context;
	std::cerr << "Not implemented" << std::endl;
}

void on_read_result(void *context, const elliptics::sync_read_result &result, const elliptics::error_info &error)
{
	if (error) {
		go_read_callback(NULL, 0, error.code(), context);
	} else {
		std::vector<go_read_result> to_go;
		for (size_t i = 0; i < result.size(); i++) {
			to_go.push_back(go_read_result{(char *)result[i].file().data()});
		}
		go_read_callback(&to_go[0], result.size(), error.code(), context);
	}
}

void on_write_result(void *context, const elliptics::sync_write_result &result, const elliptics::error_info &error)
{
	if (error) {
		go_read_callback(NULL, 0, error.code(), context);
	} else {
		std::vector<go_write_result> to_go;

		for (size_t i = 0 ; i < result.size(); i++) {
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
	return new elliptics::session(*node);
}

void session_set_groups(ell_session *session, int32_t *groups, int count)
{
	std::vector<int> g(groups, groups + count);
	session->set_groups(g);
}

void session_read_data(ell_session *session, void *context, ell_key *key)
{
	using namespace std::placeholders;
	try {
		session->read_data(*key, 0, 0).connect(std::bind(&on_read_result, context, _1, _2));
	} catch (elliptics::error &e) {
		std::cerr << e.what() << std::endl;
	}
}

void session_write_data(ell_session *session, void *context, ell_key *key, char *data, size_t size)
{
	using namespace std::placeholders;
	session->write_data(*key, elliptics::data_pointer(data, size), 0).connect(std::bind(&on_write_result, context, _1, _2));
}

void session_remove(ell_session *session, void *context, ell_key *key)
{
	using namespace std::placeholders;
	session->remove(*key).connect(std::bind(&on_remove, context, _1, _2));
}

void session_stat_log(ell_session *session, void *context)
{
	using namespace std::placeholders;
	session->stat_log().connect(std::bind(&on_stat_result, context, _1));
}

}
