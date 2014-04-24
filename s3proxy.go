package main

import (
	"flag"
	"log"
	"net/http"

	s3 "github.com/noxiouz/elliptics-go/ellipticsS3"
)

var (
	endpoint string
)

func init() {
	flag.StringVar(&endpoint, "endpoint", ":9000", "riftendpoint")

	flag.Parse()
}

func main() {
	h, err := s3.GetRouter(endpoint)
	if err != nil {
		log.Fatalf("Unable to create backend %s", err)
	}
	http.ListenAndServe(":9000", h)
}
