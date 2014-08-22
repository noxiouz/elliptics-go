#!/bin/sh
go test -v -coverprofile=coverage.out github.com/noxiouz/elliptics-go/elliptics && go tool cover -func=coverage.out && go tool cover -html=coverage.out -o cov.html
