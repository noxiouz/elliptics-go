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

package elliptics

/*
#include "key.h"
#include <stdio.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"sort"
	"unsafe"
)

var (
	InvalidKeyArgument = errors.New("InvalidKeyArgument")
)

type Key struct {
	key unsafe.Pointer
}

func NewKey(args ...interface{}) (key *Key, err error) {
	var ckey unsafe.Pointer
	if len(args) == 1 {
		if remote, ok := args[0].(string); ok {
			_cRemote := C.CString(remote)
			defer C.free(unsafe.Pointer(_cRemote))
			ckey = C.new_key_remote(_cRemote)
			if ckey == nil {
				err = fmt.Errorf("could not create new key")
				return
			}
		} else {
			err = InvalidKeyArgument
			return
		}
	} else {
		ckey = C.new_key()
		if ckey == nil {
			err = fmt.Errorf("could not create new key")
			return
		}
	}
	return &Key{ckey}, nil
}

func NewKeyFromIdStr(id string) (key *Key, err error) {
	cid := C.CString(id)
	defer C.free(unsafe.Pointer(cid))
	ckey := C.new_key_from_id(cid)
	if ckey == nil {
		err = fmt.Errorf("could not create new key")
		return
	}

	return &Key{ckey}, nil
}

func (k *Key) ById() bool {
	return int(C.key_by_id(k.key)) > 0
}

func (k *Key) SetId(id []byte, group_id uint32) {
	C.key_set_id(k.key, unsafe.Pointer(&id[0]), C.int(len(id)), C.int(group_id))
	return
}
func (k *Key) SetRawId(id []byte) {
	C.key_set_raw_id(k.key, unsafe.Pointer(&id[0]), C.int(len(id)))
	return
}

func (k *Key) Free() {
	C.delete_key(k.key)
}

func (k *Key) CmpID(id []uint8) int {
	return int(C.key_id_cmp(k.key, unsafe.Pointer(&id[0])))
}

type Keys struct {
	keys unsafe.Pointer // ell_keys structure pointer
}

func NewKeys(keys []string) (ret Keys, err error) {
	var ckeys unsafe.Pointer
	ckeys = C.ell_keys_new()
	if ckeys == nil {
		err = fmt.Errorf("could not create key array")
		return
	}

	sort.Strings(keys)
	for _, k := range keys {
		ckey := C.CString(k)
		defer C.free(unsafe.Pointer(ckey))
		C.ell_keys_insert(ckeys, ckey, C.int(len(k)))
	}

	err = nil
	ret = Keys{
		keys: ckeys,
	}

	return
}

func (kk *Keys) Find(id []uint8) (ret string, err error) {
	var tmp *C.char
	tmp = C.ell_keys_find(kk.keys, unsafe.Pointer(&id[0]))
	defer C.free(unsafe.Pointer(tmp))

	if tmp == nil {
		err = fmt.Errorf("could not find key for given ID")
		return
	}

	ret = C.GoString(tmp)
	err = nil
	return
}

func (kk *Keys) Free() {
	C.ell_keys_free(kk.keys)
}

// Key based on std::vector<dnet_raw_id>

type DnetRawIDKeys struct {
	keys unsafe.Pointer
}

func NewDnetRawIDKeys(ids []*DnetRawID) (kk *DnetRawIDKeys, err error) {
	var ckeys unsafe.Pointer
	ckeys = C.ell_dnet_raw_id_keys_new()
	if ckeys == nil {
		err = fmt.Errorf("could not create key array")
		return
	}

	err = nil
	kk = &DnetRawIDKeys {
		keys: ckeys,
	}

	for _, id := range ids {
		kk.InsertID(id)
	}

	return
}

func (kk *DnetRawIDKeys) InsertID(id *DnetRawID) {
	C.ell_dnet_raw_id_keys_insert(kk.keys, unsafe.Pointer(&id.ID[0]), C.int(len(id.ID)))
}

func (kk *DnetRawIDKeys) Free() {
	C.ell_dnet_raw_id_keys_free(kk.keys)
}

func (kk *DnetRawIDKeys) Size() int {
	return int(C.ell_dnet_raw_id_keys_size(kk.keys))
}
