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

package elliptics

/*
#include "key.h"
#include <stdio.h>
*/
import "C"

import (
	"unsafe"
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
			ckey, err = C.new_key_remote(_cRemote)
		}
	} else {
		ckey, err = C.new_key()
		if err != nil {
			return
		}
	}
	return &Key{ckey}, nil
}

func (k *Key) ById() bool {
	return int(C.key_by_id(k.key)) > 0
}

// func (k *Key) SetId(dnetId *DnetId) (err error) {
// 	_, err = C.key_set_id(k.key, dnetId._dnet_id)
// 	return
// }

func (k *Key) Free() {
	C.delete_key(k.key)
}
