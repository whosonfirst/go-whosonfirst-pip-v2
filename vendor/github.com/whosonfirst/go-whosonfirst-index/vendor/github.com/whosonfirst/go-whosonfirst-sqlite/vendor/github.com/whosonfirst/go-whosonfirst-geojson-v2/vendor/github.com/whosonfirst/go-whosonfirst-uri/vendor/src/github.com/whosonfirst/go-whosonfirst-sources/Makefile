CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	rmdeps deps fmt bin

self:   prep
	if test -d src/github.com/whosonfirst/go-whosonfirst-sources; then rm -rf src/github.com/whosonfirst/go-whosonfirst-sources; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-sources/sources
	cp sources/*.go src/github.com/whosonfirst/go-whosonfirst-sources/sources
	cp *.go src/github.com/whosonfirst/go-whosonfirst-sources/

deps:   self

fmt:
	go fmt *.go
	go fmt sources/*.go
	go fmt cmd/*.go

test:	self
	@GOPATH=$(GOPATH) go run cmd/test.go

spec:	self
	@GOPATH=$(GOPATH) go run cmd/mk-spec.go > sources/spec.go
