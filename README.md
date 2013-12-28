elliptics-go
============

Go binding for [Elliptics](https://github.com/reverbrain/elliptics) key-value storage.

## Example

``` go
package main

import (
	"flag"
	"log"
	"time"

	"github.com/noxiouz/elliptics-go/elliptics"
)

var HOST string
var KEY string

func init() {
	flag.StringVar(&HOST, "host", ELLHOST, "elliptics host:port:family")
	flag.StringVar(&KEY, "key", TESTKEY, "key")
	flag.Parse()
}

const TESTKEY = "TESTKEYsssd"
const ELLHOST = "elstorage01f.kit.yandex.net:1025:2"

func main() {
	// Create file logger
	level := 2
	EllLog, err := elliptics.NewFileLogger("/tmp/elliptics-go.log", level)
	if err != nil {
		log.Fatalln("NewFileLogger: ", err)
	}
	defer EllLog.Free()
	EllLog.Log(elliptics.INFO, "started: %v, level: %d", time.Now(), level)

	// Create elliptics node
	node, err := elliptics.NewNode(EllLog)
	if err != nil {
		log.Println(err)
	}
	defer node.Free()

	node.SetTimeouts(100, 1000)
	if err = node.AddRemote(HOST); err != nil {
		log.Fatalln("AddRemote: ", err)
	}

	session, err := elliptics.NewSession(node)
	if err != nil {
		log.Fatal("Error", err)
	}

	session.SetGroups([]int32{1, 2, 3})
	session.SetNamespace("TEST3")

	log.Println("Find all")
	for res := range session.FindAllIndexes([]string{"F", "T"}) {
		log.Printf("%v", res.Data())
	}
	log.Println("Find any")
	for res := range session.FindAnyIndexes([]string{"F", "T"}) {
		log.Printf("%v", res.Data())
	}

	for rd := range session.ReadData(KEY) {
		log.Printf("%s \n", rd.Data())
	}

	for rw := range session.WriteData(KEY, "TESTDATA") {
		log.Println(rw)
	}

	lookuped_key, _ := elliptics.NewKey(KEY)
	defer lookuped_key.Free()
	for lookUp := range session.Lookup(lookuped_key) {
		log.Println(lookUp)
	}

	indexes := map[string]string{
		"F": "indexF",
		"A": "indexA",
	}
	if si, ok := <-session.SetIndexes(KEY, indexes); !ok {
		log.Println("SetIndexes successfully")
	} else {
		log.Println("SetIndexes error: ", si.Error())
	}

	log.Println("List indexes for key ", KEY)
	for li := range session.ListIndexes(KEY) {
		log.Println("Index: ", li.Data)
	}

	indexes["TTT"] = "IndexTTT"
	if ui, ok := <-session.UpdateIndexes(KEY, indexes); !ok {
		log.Println("UpdateIndexes successfully")
	} else {
		log.Println("UpdateIndexes error: ", ui.Error())
	}

	log.Println("List indexes for key ", KEY)
	for li := range session.ListIndexes(KEY) {
		log.Println("Index: ", li.Data)
	}

	//KEY exists
	if rm, ok := <-session.Remove(KEY); !ok {
		log.Println("Remove successfully")
	} else {
		log.Println("Removing error: ", rm.Error())
	}

	//KEY doesn't exist
	if rm, ok := <-session.Remove(KEY); !ok {
		log.Println("Remove successfully")
	} else {
		log.Println("Removing error: ", rm.Error())
	}
}
```

## Installation

You should install `elliptics-client-dev` to build this one.
It could be installed from a [repository](http://repo.reverbrain.com)
or build from a [source](https://github.com/reverbrain/elliptics).
```
go get github.com/noxiouz/elliptics-go/elliptics
```

Specify the following environment variables, if libraries and heders are located in a non-standard location:

 * `CGO_CFLAGS` - C flags
 * `CGO_CPPFLAGS` - both C++/C flags
 * `CGO_CXXFLAGS` - C++ flags
 * `CGO_LDFLAGS` - linker flags

