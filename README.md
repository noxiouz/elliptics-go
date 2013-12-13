elliptics-go
============

Go binding for Elliptics key-value storage.

## Installation

```
go get github.com/noxiouz/elliptics-go/elliptics
```

Specify the following environment variables, if libraries and heders are located in a non-standard location:
 + CGO_CFLAGS - C flags
 + CGO_CPPFLAGS - both C++/C flags
 + CGO_CXXFLAGS - C++ flags
 + CGO_LDFLAGS - flags for linker

## Example

``` go
package main

import (
	"fmt"
	"github.com/noxiouz/elliptics-go/elliptics"
	"log"
	"time"
)

const TESTKEY = "TESTKEY"
const ELLHOST = "ELLIPTICSHOST:1025"

func main() {
	// Create file logger
	log.Println("Create logger")
	EllLog, err := elliptics.NewFileLogger("LOG.log")
	if err != nil {
		log.Fatalln("NewFileLogger: ", err)
	}
	defer EllLog.Free()
	EllLog.Log(4, fmt.Sprintf("%v\n", time.Now()))

	// Create elliptics node
	log.Println("Create elliptics node")
	node, err := elliptics.NewNode(EllLog)
	if err != nil {
		log.Println(err)
	}
	defer node.Free()

	node.SetTimeouts(100, 1000)
	log.Println("Add remotes")
	if err = node.AddRemote(ELLHOST); err != nil {
		log.Fatalln("AddRemote: ", err)
	}

	session, err := elliptics.NewSession(node)
	if err != nil {
		log.Fatal("Error", err)
	}
	log.Println("Session ", session)

	session.SetGroups([]int32{1, 2, 3})
	rd := <-session.ReadData(TESTKEY)
	if rd.Error() != nil {
		log.Fatal("read error ", rd.Error())
	}
	log.Printf("%s \n", rd.Data())
	rw := <-session.WriteData(TESTKEY, "dummydata")
	if rw.Error() != nil {
		log.Fatal("write error", rw.Error())
	}
	rd = <-session.ReadData(TESTKEY)
	if rd.Error() != nil {
		log.Fatal("read error ", rd.Error())
	}
	log.Printf("%s \n", rd.Data())
}

```
