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

#ifndef NODE_H
#define NODE_H

#include "logger.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef void ell_node;

ell_node*
new_node(ell_file_logger *fl);

void
delete_node(ell_node *node);

void                        
node_add_remote(ell_node *node, const char *addr, const int port, const int family);

void 
node_set_timeouts(ell_node *node, const int wait_timeout, const int check_timeout);


#ifdef __cplusplus 
}
#endif

#endif
