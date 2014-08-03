package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/noxiouz/elliptics-go/ellipticsS3"
)

var (
	riftendpoint  string
	proxyendpoint string
	metagroups    groupSliceFlag
	datagroups    groupSliceFlag
)

type groupSliceFlag []int

func (i *groupSliceFlag) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *groupSliceFlag) Set(value string) error {
	raw := strings.Split(value, ",")
	for _, item := range raw {
		tmp, err := strconv.Atoi(item)
		if err != nil {
			return fmt.Errorf("Unable to convert %s to number", item)
		}

		*i = append(*i, tmp)
	}

	return nil
}

func main() {
	flag.StringVar(&riftendpoint, "riftendpoint", ":9000", "riftendpoint")
	flag.StringVar(&proxyendpoint, "endpoint", ":9000", "proxyendpoint")
	flag.Var(&metagroups, "metagroups", "groups for metadata")
	flag.Var(&datagroups, "datagroups", "groups for data")
	flag.Parse()

	if len(metagroups) == 0 || len(datagroups) == 0 {
		log.Fatal("Please, specify data and metadata groups")
		return
	}

	cfg := ellipticsS3.Config{
		Endpoint:       riftendpoint,
		MetaDataGroups: metagroups,
		DataGroups:     datagroups,
	}

	log.Printf("Connecting to Rift %s", riftendpoint)
	h, err := ellipticsS3.GetRouter(cfg)
	if err != nil {
		log.Fatalf("Unable to create backend %s", err)
	}

	log.Printf("Listening  %s", proxyendpoint)
	err = http.ListenAndServe(proxyendpoint, h)
	if err != nil {
		log.Fatal(err)
	}
}
