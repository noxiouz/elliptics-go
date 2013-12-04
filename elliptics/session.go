package elliptics

import (
	"unsafe"
)

/*
#include "session.h"
#include <stdio.h>
*/
import "C"

type Session struct {
	session unsafe.Pointer
}

func NewSession(node *Node) (*Session, error) {
	session, err := C.new_elliptics_session(node.node)
	if err != nil {
		return nil, err
	}
	return &Session{session}, err
}

func (s *Session) SetGroups(groups []int) {
	// count := C.int(len(groups))
	// f := (*C.int)(&groups[0])
	// fmt.Println(&f, " ", &groups[0])
	//C.session_set_groups(s.session, f, count)
}

func (s *Session) Read(key *Key) (a chan bool) {
	a = make(chan bool, 1)
	context := func(result unsafe.Pointer) {
		a <- true
	}
	C.session_read_data(s.session, unsafe.Pointer(&context), key.key)
	return
}

func (s *Session) StatLog() (a chan bool) {
	a = make(chan bool, 1)
	context := func(result unsafe.Pointer) {
		a <- true
	}
	C.session_stat_log(s.session, unsafe.Pointer(&context))
	return
}
