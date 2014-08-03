.PHONY: proxy deps

all: proxy

deps:
	go get github.com/gorilla/handlers
	go get github.com/gorilla/mux

proxy: deps
	go build s3proxy.go


