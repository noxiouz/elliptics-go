package main

import (
	"flag"
	"log"

	"github.com/noxiouz/elliptics-go/rift"
)

var (
	bucketDirName string

	endpoint string = "localhost:8080"

	userBucketDirOpt = rift.BucketDirectoryOptions{
		Groups:    []int{2},
		ACL:       make([]rift.ACLStruct, 0),
		Flags:     0,
		MaxSize:   0,
		MaxKeyNum: 0,
	}
)

func main() {
	flag.StringVar(&bucketDirName, "username", "", "create bucket directory for user")
	flag.Parse()

	log.Printf("Create user %s using endpoint %s", bucketDirName, endpoint)
	r, err := rift.NewRiftClient(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	info, err := r.CreateBucketDir(bucketDirName, userBucketDirOpt)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%v", info)
}
