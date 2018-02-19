CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep
	if test -d src/github.com/whosonfirst/go-whosonfirst-uri; then rm -rf src/github.com/whosonfirst/go-whosonfirst-uri; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-uri
	cp uri.go src/github.com/whosonfirst/go-whosonfirst-uri/uri.go
	cp -r vendor/src/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

deps:   
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-sources"

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:	self
	go fmt uri.go

bin:	self
