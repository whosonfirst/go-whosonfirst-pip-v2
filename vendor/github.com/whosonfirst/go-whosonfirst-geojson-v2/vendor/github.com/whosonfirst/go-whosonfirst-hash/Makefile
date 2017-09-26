CWD=$(shell pwd)
GOPATH := $(CWD)

build:	fmt bin

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-hash
	cp *.go src/github.com/whosonfirst/go-whosonfirst-hash/
	# cp -r vendor/src/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

deps:   rmdeps

vendor-deps: deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go
	go fmt cmd/*.go

bin:	self
	@GOPATH=$(GOPATH) go build -o bin/wof-hash-file cmd/wof-hash-file.go
