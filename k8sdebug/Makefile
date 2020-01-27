.PHONY: build plugin agent check

GOENV  := GO15VENDOREXPERIMENT="1" GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GO := $(GOENV) go
GOPATH := $(GOPATH)

default: plugin

plugin:
	GO111MODULE=on go build -o kubectl-debug main.go
	cp kubectl-debug $(GOPATH)/bin/

check:
	find . -iname '*.go' -type f | grep -v /vendor/ | xargs gofmt -l
	GO111MODULE=on go test -v -race ./...
	$(GO) vet ./...
