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

/*
#cgo LDFLAGS: -lelliptics_client -lelliptics_cpp -lboost_thread -lboost_system

#include "node.h"
#include <stdlib.h>

static char**makeCharArray(int size) {
        return calloc(sizeof(char*), size);
}

static void setArrayString(char **a, char *s, int n) {
        a[n] = s;
}

static void freeCharArray(char **a, int size) {
        int i;
        for (i = 0; i < size; i++)
                free(a[i]);
        free(a);
}
*/
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

// A Node is responsible for the connection with a server side.
// Also it is responsible for checking timeouts, maintenance and checking of communication.
// To initialize the Node you should use NewNode.
type Node struct {
	node   unsafe.Pointer
}

// NewNode returns new Node with a given Logger.
func NewNode(logfile string, level string) (node *Node, err error) {
	clevel := C.CString(level)
	defer C.free(unsafe.Pointer(clevel))

	clogfile := C.CString(logfile)
	defer C.free(unsafe.Pointer(clogfile))

	cnode := C.new_node(clogfile, clevel)
	if cnode == nil {
		err = fmt.Errorf("could not create node, please check stderr output")
		return
	}
	node = &Node{cnode}
	return
}

// Free disposes given Node instance.
// Do not destroy the Node used by any Session.
func (node *Node) Free() {
	C.delete_node(node.node)
}

/*
 * SetTimeouts overrides the default values for timeouts.
 *
 * waitTimeout affects to any transaction, which is sent to the cluster.
 * Default value is 5 seconds.
 *
 * checkTimeout is responsible for updating the routing table
 * and checking the network connection.
 * By default it's 60 seconds.
 */
func (node *Node) SetTimeouts(waitTimeout int, checkTimeout int) {
	C.node_set_timeouts(node.node, C.int(waitTimeout), C.int(checkTimeout))
}

/*
 * AddRemote adds a connection to elliptics servers.
 *
 * Address is specified as Host:Port:Family. Family can be omitted.
 * Suitable Family values are: 2 (AF_INET) and 10 (AF_INET6).
 */
func (node *Node) AddRemote(addr string) (err error) {
	caddr := C.CString(addr)
	defer C.free(unsafe.Pointer(caddr))

	err = nil
	c_err := C.node_add_remote_one(node.node, caddr)
	if c_err < 0 {
		err = syscall.Errno(-c_err)
	}
	return
}

/*
 * AddRemotes adds a connection to elliptics servers, it runs in parallel
 * and connects to multiple nodes at once.
 *
 * Address is specified as Host:Port:Family. Family can be omitted.
 * Suitable Family values are: 2 (AF_INET) and 10 (AF_INET6).
 */

func (node *Node) AddRemotes(addrs []string) (err error) {
	if len(addrs) == 0 {
		return fmt.Errorf("list of remotes is empty")
	}

	cargs := C.makeCharArray(C.int(len(addrs)))
	defer C.freeCharArray(cargs, C.int(len(addrs)))
	for i, s := range addrs {
		C.setArrayString(cargs, C.CString(s), C.int(i))
	}

	err = nil
	c_err := C.node_add_remote_array(node.node, cargs, C.int(len(addrs)))
	if c_err < 0 {
		err = syscall.Errno(-c_err)
	}
	return
}
