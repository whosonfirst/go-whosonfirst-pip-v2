CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-sqlite
	cp -r database src/github.com/whosonfirst/go-whosonfirst-sqlite/
	cp -r index src/github.com/whosonfirst/go-whosonfirst-sqlite/
	cp -r tables src/github.com/whosonfirst/go-whosonfirst-sqlite/
	cp -r utils src/github.com/whosonfirst/go-whosonfirst-sqlite/
	cp -r *.go src/github.com/whosonfirst/go-whosonfirst-sqlite/
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/mattn/go-sqlite3"
	@GOPATH=$(GOPATH) go install "github.com/mattn/go-sqlite3"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-spatialite"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-flags"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-index"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-log"
	rm -rf src/github.com/whosonfirst/go-whosonfirst-index/vendor/github.com/whosonfirst/go-whosonfirst-sqlite/

vendor-deps: rmdeps deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go
	go fmt database/*.go
	go fmt index/*.go
	go fmt tables/*.go
	go fmt utils/*.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/wof-sqlite-index-example cmd/wof-sqlite-index-example.go