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

#ifndef __ELLIPTICS_KEY_H
#define __ELLIPTICS_KEY_H


#ifdef __cplusplus
#include <elliptics/session.hpp>
typedef ioremap::elliptics::key ell_key;

typedef struct
{
	std::vector<ioremap::elliptics::key>	kk;
	void insert(const char *str, int len) {
		std::string tmp(str, len);
		kk.emplace_back(std::move(ioremap::elliptics::key(tmp)));
	}
	std::string find(const ioremap::elliptics::key &tmp) const {
		auto it = std::lower_bound(kk.begin(), kk.end(), tmp, *this);
		if (it != kk.end()) {
			return it->remote();
		}

		return std::string();
	}
	// we can not use elliptics::key::operator< here, since it checks
	// whether operands were both created using only id, but in our
	// case key stored in @kk array was created from string (not id),
	// and key to be found was created from id
	bool operator()(const ioremap::elliptics::key &k1, const ioremap::elliptics::key &k2) const {
		return memcmp(&k1.id().id, &k2.id().id, DNET_ID_SIZE) < 0;
	}
} ell_keys;

extern "C" {
#else
#include <elliptics/interface.h>
typedef void ell_key;
typedef void ell_keys;
#endif

ell_key *new_key();
ell_key *new_key_remote(const char *remote);

const char *key_remote(ell_key *key);

int key_by_id(ell_key *key);
void key_set_id(ell_key *c_key, const struct dnet_id *id);
int key_id_cmp(ell_key *key, const void *id);

void delete_key(ell_key *c_key);

ell_keys *ell_keys_new();
void ell_keys_free(ell_keys *keys);
int ell_keys_insert(ell_keys *keys, const char *str, int len);
char *ell_keys_find(ell_keys *keys, void *id);

#ifdef __cplusplus
}
#endif

#endif /* __ELLIPTICS_KEY_H */
