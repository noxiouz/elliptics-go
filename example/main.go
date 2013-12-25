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
	for res := range session.FindAllIndexes([]string{"G", "Z", "Y", "T"}) {
		log.Printf("%v", res.Data())
	}
	log.Println("Find any")
	for res := range session.FindAnyIndexes([]string{"G", "Z", "Y", "T"}) {
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
