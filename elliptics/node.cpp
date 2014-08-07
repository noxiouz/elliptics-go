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

ell_node *new_node(ell_file_logger *fl)
{
    //ToDO: attach logger
	elliptics::node *node = new elliptics::node();
	return node;
}

void delete_node(ell_node *node)
{
	delete node;
}

int node_add_remote(ell_node *node, const char *addr, const int port, const int family)
{
	try {
        auto address = elliptics::address(addr, port, family);
		node->add_remote(address);
	} catch(const elliptics::error &e) {
		return e.error_code();
	}

	return 0;
}

int node_add_remote_one(ell_node *node, const char *addr)
{
	try {
		node->add_remote(addr);
	} catch(const elliptics::error &e) {
		return e.error_code();
	}

	return 0;
}

int node_add_remote_array(ell_node *node, const char **addr, const int num)
{
	try {
        std::vector<elliptics::address> vaddr;
        vaddr.reserve(num);
        for (int i = 0; i < num; i++) {
            vaddr.push_back(elliptics::address(*(addr + num)));
        }
		node->add_remote(vaddr);
	} catch(const elliptics::error &e) {
		return e.error_code();
	}
	return 0;
}


void node_set_timeouts(ell_node *node, const int wait_timeout, const int check_timeout)
{
	node->set_timeouts(wait_timeout, check_timeout);
}

} // extern "C"
