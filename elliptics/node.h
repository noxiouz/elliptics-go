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

#ifndef __ELLIPTICS_NODE_H
#define __ELLIPTICS_NODE_H

#ifdef __cplusplus
#include <elliptics/session.hpp>

#define BOOST_BIND_NO_PLACEHOLDERS
#include <blackhole/formatter/string.hpp>
#undef BOOST_BIND_NO_PLACEHOLDERS

class go_logger_base: public ioremap::elliptics::logger_base {
public:
	go_logger_base(void *priv, const char *level);
	std::string format();
};

class ell_node : public ioremap::elliptics::node {
public:
	ell_node(std::shared_ptr<go_logger_base> &base, dnet_config &cfg) :
		::ioremap::elliptics::node(
			ioremap::elliptics::logger(*base,
				blackhole::log::attributes_t({
					ioremap::elliptics::keyword::request_id() = 0
				})
			),
		cfg),
		m_log(base) {
	}

private:
	std::shared_ptr<go_logger_base> m_log;
};

extern "C" {
#else
typedef void ell_node;
#endif

ell_node *new_node(void *priv, const char *level);
void delete_node(ell_node *node);

int node_add_remote(ell_node *node, const char *addr, int port, int family);
int node_add_remote_one(ell_node *node, const char *addr);
int node_add_remote_array(ell_node *node, const char **addr, int num);

void node_set_timeouts(ell_node *node, int wait_timeout, int check_timeout);

#ifdef __cplusplus
}
#endif

#endif
