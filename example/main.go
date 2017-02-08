package main

import (
	"log"
	"os"
	"strings"

	"github.com/noxiouz/elliptics-go/elliptics"
)

func main() {
	node, err := elliptics.NewNode(os.DevNull, "error")
	if err != nil {
		log.Fatalf("failed to create node %v", err)
	}
	node.AddRemote("localhost:1024:2")
	defer node.Free()

	session, err := elliptics.NewSession(node)
	if err != nil {
		log.Fatalf("failed to create Session %v", err)
	}
	defer session.Delete()

	session.SetGroups([]uint32{1, 2, 3})
	session.SetCflags(elliptics.DNET_FLAGS_NOCACHE)
	session.SetIOflags(elliptics.DNET_IO_FLAGS_NOCSUM)

	const (
		nm  = "examplenamespace"
		key = "examplekey"
	)
	var data = strings.NewReader("exampledata")
	session.SetNamespace(nm)
	for res := range session.WriteData(key, data, 0, 0) {
		if err := res.Error(); err != nil {
			log.Fatalf("WriteData returns error %v", err)
		}
	}

	for res := range session.ReadData(key, 0, 0) {
		if res.Error() == nil {
			log.Printf("%s", res.Data())
		} else {
			log.Printf("ReadData error %v", res.Error())
		}
	}

	for res := range session.Remove(key) {
		if err := res.Error(); err != nil {
			log.Printf("Remove error %v", err)
		}
	}
}
