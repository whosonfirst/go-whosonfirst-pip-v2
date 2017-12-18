CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	rmdeps deps fmt bin

self:   prep
	if test -d src/github.com/whosonfirst/go-whosonfirst-placetypes; then rm -rf src/github.com/whosonfirst/go-whosonfirst-placetypes; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-placetypes/filter
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-placetypes/placetypes
	cp *.go src/github.com/whosonfirst/go-whosonfirst-placetypes/
	cp *.go src/github.com/whosonfirst/go-whosonfirst-placetypes/
	cp filter/*.go src/github.com/whosonfirst/go-whosonfirst-placetypes/filter/
	cp placetypes/*.go src/github.com/whosonfirst/go-whosonfirst-placetypes/placetypes/

deps:   self

fmt:
	go fmt *.go
	go fmt placetypes/*.go
	go fmt filter/*.go

test:	self
	@GOPATH=$(GOPATH) go run cmd/test.go

spec:
	@GOPATH=$(GOPATH) go run cmd/mk-spec.go > placetypes/spec.go
