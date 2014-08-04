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
#define NODE_H

#include "logger.h"

#ifdef __cplusplus
#include <elliptics/session.hpp>
typedef ioremap::elliptics::node ell_node;
extern "C" {
#else
typedef void ell_node;
#endif

ell_node *new_node(ell_file_logger *fl);
void delete_node(ell_node *node);

int node_add_remote(ell_node *node, const char *addr, const int port, const int family);
int node_add_remote_one(ell_node *node, const char *addr);
int node_add_remote_array(ell_node *node, const char **addr, const int num);

void node_set_timeouts(ell_node *node, const int wait_timeout, const int check_timeout);

#ifdef __cplusplus
}
#endif

#endif
