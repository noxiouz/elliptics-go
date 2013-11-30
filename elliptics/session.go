package elliptics

import (
	"fmt"
	"unsafe"
)

/*
  #cgo LDFLAGS: -lell -lelliptics_cpp -L .
  #include "lib/session.h"

 	extern void Test(void*, void*);

	static void my_Test(void *s, void *ch) {
		session_stat_log(s, &Test, ch);
	}
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

func (s *Session) Read(key *Key) {
	C.session_read_data(s.session, key.key)
}

func (s *Session) StatLog() (a chan bool) {
	//C.session_stat_log(s.session)
	a = make(chan bool, 1)
	fmt.Println("Pointer tot channel", &a)
	C.my_Test(s.session, unsafe.Pointer(&a))
	return
}

//export Test
func Test(p unsafe.Pointer, ch unsafe.Pointer) {
	fmt.Println("CALLBACK CALLBACK mypointer", p, ch)
	_ch := (*chan bool)(ch)
	*_ch <- true
}
