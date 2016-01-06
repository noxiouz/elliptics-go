/*
* 2013+ Copyright (c) Anton Tyurin <noxiouz@yandex.ru>
* All rights reserved.
*
* This program is free software; you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation; either version 3 of the License, or
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

ell_node *new_node(const char *logfile, const char *level)
{
	try {
		dnet_config cfg;
		memset(&cfg, 0, sizeof(dnet_config));
		//cfg.flags = DNET_CFG_MIX_STATES;
		cfg.io_thread_num = 8;
		cfg.nonblocking_io_thread_num = 4;
		cfg.net_thread_num = 4;
		cfg.wait_timeout = 15;
		cfg.check_timeout = 30;

		std::shared_ptr<elliptics::file_logger> base =
			std::make_shared<elliptics::file_logger>(logfile, elliptics::file_logger::parse_level(level));

		return new ell_node(base, cfg);
	} catch (const std::exception &e) {
		fprintf(stderr, "could not create new node: exception: %s\n", e.what());
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
		std::vector<ioremap::elliptics::address> vaddr;
		for (int i = 0; i < num; ++i) {
			try {
				vaddr.push_back(ioremap::elliptics::address(addr[i]));
			} catch (...) {
				// we do not care if it failed to create address
			}

		}
		node->add_remote(vaddr);
	} catch (const elliptics::error &e) {
		return e.error_code();
	} catch (const std::exception &e) {
		return -EINVAL;
	}
	return 0;
}

void node_set_timeouts(ell_node *node, int wait_timeout, int check_timeout)
{
	node->set_timeouts(wait_timeout, check_timeout);
}

} // extern "C"
