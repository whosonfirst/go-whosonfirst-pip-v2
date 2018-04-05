CWD=$(shell pwd)
GOPATH := $(CWD)

build:	fmt bin

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-geojson-v2/
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-geojson-v2/feature
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-geojson-v2/properties
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-geojson-v2/utils
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-geojson-v2/whosonfirst
	cp -r *.go src/github.com/whosonfirst/go-whosonfirst-geojson-v2/
	cp -r feature/*.go src/github.com/whosonfirst/go-whosonfirst-geojson-v2/feature/
	cp -r geometry/*.go src/github.com/whosonfirst/go-whosonfirst-geojson-v2/geometry/
	cp -r properties/* src/github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/
	cp -r utils/*.go src/github.com/whosonfirst/go-whosonfirst-geojson-v2/utils/
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

deps:   rmdeps
	@GOPATH=$(GOPATH) go get -u "github.com/skelterjohn/geom"
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/gjson"
	@GOPATH=$(GOPATH) go get -u "github.com/mmcloughlin/geohash"
	# @GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-export"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-flags"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-hash"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-placetypes"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-spr"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-uri"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/warning"

vendor-deps: deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go
	go fmt feature/*.go
	go fmt geometry/*.go
	go fmt properties/geometry/*.go
	go fmt properties/whosonfirst/*.go
	go fmt utils/*.go
	go fmt *.go

bin:	self
	@GOPATH=$(GOPATH) go build -o bin/wof-feature-to-spr cmd/wof-feature-to-spr.go
	@GOPATH=$(GOPATH) go build -o bin/wof-geojson-dump cmd/wof-geojson-dump.go
	@GOPATH=$(GOPATH) go build -o bin/wof-geojson-existential cmd/wof-geojson-existential.go
	@GOPATH=$(GOPATH) go build -o bin/wof-geojson-hash cmd/wof-geojson-hash.go
	@GOPATH=$(GOPATH) go build -o bin/wof-geojson-intersects cmd/wof-geojson-intersects.go
	@GOPATH=$(GOPATH) go build -o bin/wof-geojson-names cmd/wof-geojson-names.go
