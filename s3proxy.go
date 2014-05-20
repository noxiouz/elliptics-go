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
	endpoint   string
	metagroups groupSliceFlag
	datagroups groupSliceFlag
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
	flag.StringVar(&endpoint, "endpoint", ":9000", "riftendpoint")
	flag.Var(&metagroups, "metagroups", "groups for metadata")
	flag.Var(&datagroups, "datagroups", "groups for data")
	flag.Parse()

	if len(metagroups) == 0 || len(datagroups) == 0 {
		log.Fatal("Please, specify data and metadata groups")
		return
	}

	cfg := ellipticsS3.Config{
		Endpoint:       endpoint,
		MetaDataGroups: metagroups,
		DataGroups:     datagroups,
	}

	h, err := ellipticsS3.GetRouter(cfg)
	if err != nil {
		log.Fatalf("Unable to create backend %s", err)
	}
	http.ListenAndServe(":9000", h)
}
