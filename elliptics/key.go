package elliptics

/*
#include "key.h"
*/
import "C"

import (
	"unsafe"
)

type Key struct {
	key unsafe.Pointer
}

func NewKey() (key *Key, err error) {
	ckey, err := C.new_key()
	if err != nil {
		return
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
