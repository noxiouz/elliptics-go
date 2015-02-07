.PHONY: fmt vet lint test cover


fmt:
	test -z "$$(gofmt -s -l elliptics/*.go )" || echo "+ please format Go code with 'gofmt -s'"

vet:
	go vet ./...

lint:
	test -z "$$(golint ./...)"

test:
	go test -v -coverprofile=coverage.out github.com/noxiouz/elliptics-go/elliptics 	

cover:
	go tool cover -func=coverage.out
