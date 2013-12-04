package elliptics

/*
#include <elliptics/interface.h>
*/
import "C"

type DnetId struct {
	_dnet_id *C.struct_dnet_id
}
