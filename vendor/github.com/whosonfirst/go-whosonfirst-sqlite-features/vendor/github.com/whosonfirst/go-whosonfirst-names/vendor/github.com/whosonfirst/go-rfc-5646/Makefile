CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-rfc-5646
	mkdir -p src/github.com/whosonfirst/go-rfc-5646/tags
	mkdir -p src/github.com/whosonfirst/go-rfc-5646/subtags
	cp *.go src/github.com/whosonfirst/go-rfc-5646
	cp tags/*.go src/github.com/whosonfirst/go-rfc-5646/tags
	cp subtags/*.go src/github.com/whosonfirst/go-rfc-5646/subtags
	# cp -r vendor/src/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	# @GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-log"

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go
	go fmt cmd/*.go
	go fmt subtags/*.go
	go fmt tags/*.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/rfc-5646-parse cmd/rfc-5646-parse.go
