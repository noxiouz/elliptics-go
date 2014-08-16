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

#include "node.h"
#include <errno.h>

using namespace ioremap;

extern "C" {
#include "_cgo_export.h"

class go_logger_frontend : public blackhole::base_frontend_t
{
	public:
		go_logger_frontend(void *priv) : m_priv(priv), m_formatter("%(message)s %(...L)s") {
		}

		virtual void handle(const blackhole::log::record_t &record) {
			dnet_log_level level = record.extract<dnet_log_level>(blackhole::keyword::severity<dnet_log_level>().name());
			printf("level: %d, msg: %s\n", level, m_formatter.format(record).c_str());
		}

	private:
		void *m_priv;
		blackhole::formatter::string_t m_formatter;
};

go_logger_base::go_logger_base(void *priv, const char *level)
{
	verbosity(elliptics::file_logger::parse_level(level));

	add_frontend(blackhole::utils::make_unique<go_logger_frontend>(priv));
}

std::string go_logger_base::format()
{
	return "%(timestamp)s %(request_id)s/%(lwp)s/%(pid)s %(severity)s: %(message)s, attrs: [%(...L)s]";
}

ell_node *new_node(void *priv, const char *level)
{
	try {
		std::shared_ptr<go_logger_base> base = std::make_shared<go_logger_base>(priv, level);

		return new ell_node(base);
	} catch (...) {
		return NULL;
	}
}

void delete_node(ell_node *node)
{
	delete node;
}

int node_add_remote(ell_node *node, const char *addr, int port, int family)
{
	try {
		node->add_remote(ioremap::elliptics::address(addr, port, family));
	} catch(const elliptics::error &e) {
		return e.error_code();
	}

	return 0;
}

int node_add_remote_one(ell_node *node, const char *addr)
{
	try {
		node->add_remote(ioremap::elliptics::address(addr));
	} catch(const elliptics::error &e) {
		return e.error_code();
	}

	return 0;
}

int node_add_remote_array(ell_node *node, const char **addr, int num)
{
	try {
		std::vector<ioremap::elliptics::address> vaddr(addr, addr + num);
		node->add_remote(vaddr);
	} catch(const elliptics::error &e) {
		return e.error_code();
	}

	return 0;
}

void node_set_timeouts(ell_node *node, int wait_timeout, int check_timeout)
{
	node->set_timeouts(wait_timeout, check_timeout);
}

} // extern "C"
