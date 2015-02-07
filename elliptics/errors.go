/*
 * 2014+ Copyright (c) Evgeniy Polyakov <zbr@ioremap.net>
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
#include "session.h"
#include <stdio.h>
*/
import "C"

import (
	"fmt"
)

type DnetError struct {
	Code    int
	Flags   uint64
	Message string
}

func (err *DnetError) Error() string {
	return fmt.Sprintf("elliptics error: %d: %s", err.Code, err.Message)
}

func DnetErrorFromError(err error) *DnetError {
	if ke, ok := err.(*DnetError); ok {
		return ke
	} else {
		return nil
	}
}

func ErrorData(err error) string {
	if ke, ok := err.(*DnetError); ok {
		return ke.Message
	}

	return err.Error()
}

func ErrorStatus(err error) int {
	if ke, ok := err.(*DnetError); ok {
		return ke.Code
	}

	return -22 // -EINVAL
}
