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

#include "key.h"


using namespace ioremap;

extern "C" {

ell_key* new_key()
{
    return new elliptics::key();
}

ell_key* new_key_remote(const char * remote)
{
	return new elliptics::key(std::string(remote));
}

const char* key_remote(ell_key *key)
{
	std::string remote(key->remote());
	return remote.c_str();
}

int key_by_id(ell_key *key)
{
	return key->by_id();
}

void key_set_id(ell_key *key, const struct dnet_id *id)
{
	key->set_id(*id);
}

void delete_key(ell_key *key)
{
	delete key;
}

} // extern "C"
 
