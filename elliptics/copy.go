/*
 * 2014+ copyright (c) evgeniy polyakov <zbr@ioremap.net>
 * all rights reserved.
 *
 * this program is free software; you can redistribute it and/or modify
 * it under the terms of the gnu general public license as published by
 * the free software foundation; either version 2 of the license, or
 * (at your option) any later version.
 *
 * this program is distributed in the hope that it will be useful,
 * but without any warranty; without even the implied warranty of
 * merchantability or fitness for a particular purpose. see the
 * gnu general public license for more details.
 */

package elliptics

import (
	"reflect"
	"unsafe"
)

/*
#include "stat.h"
*/
import "C"


func CopyBytes(arr unsafe.Pointer, length int) []byte {
	hdr := reflect.SliceHeader {
			Data: uintptr(arr),
			Len:  length,
			Cap:  length,
		}
	tmp := *(*[]byte)(unsafe.Pointer(&hdr))

	ret := make([]byte, length)
	copy(ret, tmp)

	return ret
}
