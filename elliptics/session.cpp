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

using namespace ioremap;

extern "C" {

void on_stat_result(Callback clb, void *ch, const elliptics::sync_stat_result &result) {
	std::cerr << "on_stat_result";
	//clb(result[0].statistics(), ch);
}

ell_session*
new_elliptics_session(ell_node* node) {
	return new elliptics::session(*node);
}

void
session_read_data(ell_session *session, ell_key *key) {
	//session->read_data(*key, 0, 0).connect(std::bind(&k, std::placeholders::_1));
}

void
session_stat_log(ell_session *session, Callback clb, void *ch) {
	//session->stat_log().connect(std::bind(&on_stat_result, clb, ch, std::placeholders::_1));
}

}