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

func (k *Key) SetId(dnetId *DnetId) (err error) {
	_, err = C.key_set_id(k.key, dnetId._dnet_id)
	return
}

func (k *Key) Free() {
	C.delete_key(k.key)
}
