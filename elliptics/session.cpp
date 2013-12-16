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


void on_stat_result(void *context, const elliptics::sync_stat_result &result) {
	std::cerr << "Not implemented" << std::endl;
}

void on_read_result(void *context, 
	const elliptics::sync_read_result &result,
	const elliptics::error_info &error) {
	go_read_result *to_go;
	if (error) {
		go_read_callback(NULL, 0, error.code(), context);
	} else {
		to_go = new go_read_result[result.size()];
		for (size_t i = 0; i < result.size(); i++) {
			std::string s = result[0].file().to_string();
			to_go[i].file = s.c_str();
		}
		go_read_callback(to_go, result.size(), error.code(), context);
		delete[] to_go; 
	}
}

void on_write_result(void *context, 
	const elliptics::sync_write_result &result,
	const elliptics::error_info &error) {
	go_write_result *to_go = new go_write_result[result.size()];
	for (size_t i = 0 ; i < result.size(); i++) {
		to_go[i].addr = result[i].storage_address();
		to_go[i].info = result[i].file_info();
		to_go[i].path = result[i].file_path();
	}
	go_write_callback(to_go, result.size(), error.code(), context);
	delete[] to_go;
}


ell_session*
new_elliptics_session(ell_node* node) {
	return new elliptics::session(*node);
}

void
session_set_groups(ell_session *session, int32_t *groups, int count) {
	std::vector<int> g(groups, groups + count);
	session->set_groups(g);
	std::cerr << "Setup " << session->get_groups().size() << " groups" << std::endl;
}

void
session_read_data(ell_session *session, void *context, ell_key *key) {
	using namespace std::placeholders;
	try {
		session->read_data(*key, 0, 0).connect(std::bind(&on_read_result,
			context,
			_1, _2));
	} catch (elliptics::error &e) {
		std::cerr << e.what() << std::endl;
	}
}

void
session_write_data(ell_session *session, void *context, 
	ell_key *key, 
	char *data,
	size_t size) {
	using namespace std::placeholders;
	session->write_data(*key, elliptics::data_pointer(data, size), 0).connect(std::bind(&on_write_result,
		context,
		_1, _2));
}

void
session_stat_log(ell_session *session, void *context) {
	using namespace std::placeholders;
	session->stat_log().connect(std::bind(&on_stat_result,
		context,
		_1));
}

}