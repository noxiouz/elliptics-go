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

#include "key.h"

using namespace ioremap;

extern "C" {

ell_key *new_key()
{
	try {
		return new elliptics::key();
	} catch (...) {
		return NULL;
	}
}

ell_key *new_key_remote(const char *remote)
{
	try {
		return new elliptics::key(std::string(remote));
	} catch (...) {
		return NULL;
	}
}

ell_key *new_key_from_id(const char *id)
{
	struct dnet_raw_id raw;
	int err;

	err = dnet_parse_numeric_id(id, raw.id);
	if (err)
		return NULL;

	try {
		return new elliptics::key(raw);
	} catch (...) {
		return NULL;
	}
}

const char *key_remote(ell_key *key)
{
	std::string remote(key->remote());
	return remote.c_str();
}

int key_by_id(ell_key *key)
{
	return key->by_id();
}

void key_set_id(ell_key *key, const void *raw, int size, int group_id)
{
	struct dnet_id id;
	memset(&id, 0, sizeof(struct dnet_id));

	memcpy(id.id, raw, size);
	id.group_id = group_id;

	key->set_id(id);
}

void key_set_raw_id(ell_key *key, const void *raw, int size)
{
	struct dnet_raw_id id;
	memset(&id, 0, sizeof(struct dnet_raw_id));

	memcpy(id.id, raw, size);

	key->set_id(id);
}

int key_id_cmp(ell_key *key, const void *id) {
	//printf("cmp: %s vs %s\n", dnet_dump_id(&key->id()), dnet_dump_id_str((const unsigned char *)id));
	return memcmp(key->raw_id().id, id, DNET_ID_SIZE);
}

void delete_key(ell_key *key)
{
	delete key;
}

ell_keys *ell_keys_new()
{
	try {
		return new ell_keys;
	} catch (...) {
		return NULL;
	}
}

int ell_keys_insert(ell_keys *keys, const char *str, int len)
{
	try {
		keys->insert(str, len);
		return 0;
	} catch (...) {
		return -ENOMEM;
	}
}

char *ell_keys_find(ell_keys *keys, void *id)
{
	try {
		dnet_raw_id raw;
		memcpy(raw.id, id, DNET_ID_SIZE);
		elliptics::key tmp(raw);
		std::string res = keys->find(tmp);
		if (res.size() == 0) {
			return NULL;
		}

		return strdup(res.c_str());
	} catch (...) {
		return NULL;
	}
}

void ell_keys_free(ell_keys *keys)
{
	delete keys;
}



ell_dnet_raw_id_keys *ell_dnet_raw_id_keys_new()
{
	try {
		return new ell_dnet_raw_id_keys;
	} catch (...) {
		return NULL;
	}
}

int ell_dnet_raw_id_keys_insert(ell_dnet_raw_id_keys *keys, const void *id, int len)
{
	try {
		keys->insert(id, len);
		return 0;
	} catch (...) {
		return -ENOMEM;
	}
}

void ell_dnet_raw_id_keys_free(ell_dnet_raw_id_keys *keys)
{
	delete keys;
}


size_t ell_dnet_raw_id_keys_size(ell_dnet_raw_id_keys *keys)
{
	return keys->ids.size();
}

} // extern "C"
