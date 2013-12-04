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
	std::cerr << "on_stat_result" << std::endl;
	GoCallback(result[0].statistics(), context);
}

void on_read_result(void *context, const elliptics::sync_read_result &result, const ioremap::elliptics::error_info &error) {
	std::cerr << "on_read_result" << std::endl;
}


ell_session*
new_elliptics_session(ell_node* node) {
	return new elliptics::session(*node);
}

void
session_set_groups(ell_session *session, int* groups, int count) {
	std::vector<int> g(groups, groups + count);
	session->set_groups(g);
}

void
session_read_data(ell_session *session, void *context, ell_key *key) {
	using namespace std::placeholders;
	std::cerr << "session_read_data DISABLED" << std::endl;
	try {
		session->read_data(*key, 0, 0).connect(std::bind(&on_read_result,
			context,
			_1, _2));
	} catch (elliptics::error &e) {
		std::cerr << e.what() << std::endl;
	}
}

void
session_stat_log(ell_session *session, void *context) {
	using namespace std::placeholders;
	session->stat_log().connect(std::bind(&on_stat_result,
		context,
		_1));
}

}