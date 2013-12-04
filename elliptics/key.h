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

#ifndef KEY_H
#define KEY_H


#ifdef __cplusplus
	#include <elliptics/session.hpp>
	typedef ioremap::elliptics::key ell_key;
	extern "C" {
#else
	#include <elliptics/interface.h>
	typedef void ell_key;
#endif

ell_key* 
new_key();

ell_key* 
new_key_remote(const char* remote);

const char* 
key_remote(ell_key * key);

int
key_by_id(ell_key * key);

void 
key_set_id(ell_key *c_key, const struct dnet_id *id);

void 
delete_key(ell_key *c_key);


#ifdef __cplusplus
}
#endif

#endif  // KEY_H
