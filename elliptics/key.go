package elliptics

// #cgo LDFLAGS: -lell -lelliptics_cpp -L .
// #include "key.h"
// #include <elliptics/interface.h>
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
