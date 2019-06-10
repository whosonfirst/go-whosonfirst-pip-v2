CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-catalog
	cp *.go src/github.com/whosonfirst/go-whosonfirst-catalog
	if test -d vendor; then cp -r vendor/* src/; fi

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:

vendor-deps: rmdeps deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go

bin: 	self
