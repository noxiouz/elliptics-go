package elliptics

/*
#cgo LDFLAGS: -lell -lelliptics_cpp -L .
#include "lib/key.h"
#include <elliptics/interface.h>

extern void my_callback(void*);

static void my_test(void *key) {
	my_callback(key);
}
*/
import "C"

import (
	"fmt"
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

func (k *Key) Call() {
	C.my_test(k.key)
}

//export  my_callback
func my_callback(p unsafe.Pointer) {
	fmt.Println("GOPA")
}
