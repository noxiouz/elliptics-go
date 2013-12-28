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

// #include "node.h"
// #include <stdlib.h>
import "C"

import (
	"syscall"
	"unsafe"
)

//A Node is responsible for the connection with the server part.
type Node struct {
	logger *Logger
	node   unsafe.Pointer
}

func isError(errno syscall.Errno) bool {
	return errno != syscall.EINPROGRESS &&
		errno != syscall.EAGAIN &&
		errno != syscall.EALREADY &&
		errno != syscall.EISCONN
}

//NewNode returns new Node given Logger.
func NewNode(log *Logger) (node *Node, err error) {
	cnode, err := C.new_node(log.logger)
	if err != nil {
		return
	}
	node = &Node{log, cnode}
	return
}

//Free disposes given Node instance.
//Do not destroy the Node used by any Session.
func (node *Node) Free() {
	C.delete_node(node.node)
}

/*SetTimeouts overrides the default values for timeouts.

waitTimeout affects to any transaction, which is sent to the cluster.
Default value is 5 seconds.

checkTimeout is responsible for updating the routing table
and checking the network connection.
By default it's 60 seconds.
*/
func (node *Node) SetTimeouts(waitTimeout int, checkTimeout int) {
	C.node_set_timeouts(node.node, C.int(waitTimeout), C.int(checkTimeout))
}

/*AddRemote adds a connection to elliptics servers.

Address is specified as Host:Port:Family. Family can be omitted.
Suitable Family values are: 2 (AF_INET) and 10 (AF_INET6).
*/
func (node *Node) AddRemote(addr string) (err error) {
	caddr := C.CString(addr)
	defer C.free(unsafe.Pointer(caddr))

	_, c_err := C.node_add_remote_one(node.node, caddr)
	if c_err != nil {
		if err, ok := c_err.(syscall.Errno); ok && isError(err) {
			return err
		} else if !ok {
			return nil
		}
	}
	return
}
