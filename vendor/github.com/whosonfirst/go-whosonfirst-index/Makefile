CWD=$(shell pwd)
GOPATH := $(CWD)

build:	rmdeps deps fmt bin

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep
	if test -d src/github.com/whosonfirst/go-whosonfirst-index; then rm -rf src/github.com/whosonfirst/go-whosonfirst-index; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-index
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-index/utils
	cp -r *.go src/github.com/whosonfirst/go-whosonfirst-index/
	cp -r utils/*.go src/github.com/whosonfirst/go-whosonfirst-index/utils/
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-crawl"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-csv"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-log"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-uri"

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	cp -r src/ vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go
	go fmt utils/*.go
	go fmt *.go

bin:	self
	@GOPATH=$(GOPATH) go build -o bin/wof-index-count cmd/wof-index-count.go	
