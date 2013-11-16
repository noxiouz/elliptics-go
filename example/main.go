package main

import (
	"fmt"
	"github.com/noxiouz/elliptics-go/elliptics"
	"log"
	"time"
)

func main() {
	// Create file logger
	l, err := elliptics.NewFileLogger("LOG.log")
	if err != nil {
		log.Fatalln("NewFileLogger: ", err)
	}
	defer l.Free()
	l.Log(4, fmt.Sprintf("%v\n", time.Now()))
	log.Println("Log level: ", l.GetLevel())

	// Create elliptics node
	node, err := elliptics.NewNode(l)
	if err != nil {
		log.Println(err)
	}

	node.SetTimeouts(100, 1000)
	if err = node.AddRemote("elstorage01f.kit.yandex.net:1025"); err != nil {
		log.Fatalln("AddRemote: ", err)
	}

}
