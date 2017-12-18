CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-names
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-names/tags
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-names/utils
	cp *.go src/github.com/whosonfirst/go-whosonfirst-names
	cp tags/*.go src/github.com/whosonfirst/go-whosonfirst-names/tags
	cp utils/*.go src/github.com/whosonfirst/go-whosonfirst-names/utils
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-rfc-5646"

vendor-deps: rmdeps deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go
	go fmt *.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/wof-names-parse cmd/wof-names-parse.go
