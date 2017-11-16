CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src/github.com/whosonfirst/go-whosonfirst-pip; then rm -rf src/github.com/whosonfirst/go-whosonfirst-pip; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-pip
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-pip/app
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-pip/cache
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-pip/filter
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-pip/http
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-pip/index
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-pip/utils
	cp *.go src/github.com/whosonfirst/go-whosonfirst-pip/
	cp -r app src/github.com/whosonfirst/go-whosonfirst-pip/
	cp -r cache src/github.com/whosonfirst/go-whosonfirst-pip/
	cp -r filter src/github.com/whosonfirst/go-whosonfirst-pip/
	cp -r flags src/github.com/whosonfirst/go-whosonfirst-pip/
	cp -r http src/github.com/whosonfirst/go-whosonfirst-pip/
	cp -r index src/github.com/whosonfirst/go-whosonfirst-pip/
	cp -r utils src/github.com/whosonfirst/go-whosonfirst-pip/
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

# if you're wondering about the 'rm -rf' stuff below it's because Go is
# weird... https://vanduuren.xyz/2017/golang-vendoring-interface-confusion/
# (20170912/thisisaaronland)

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/dhconnelly/rtreego"
	@GOPATH=$(GOPATH) go get -u "github.com/facebookgo/grace/gracehttp"
	@GOPATH=$(GOPATH) go get -u "github.com/hashicorp/golang-lru"
	@GOPATH=$(GOPATH) go get -u "github.com/jteeuwen/go-bindata/"
	@GOPATH=$(GOPATH) go get -u "github.com/elazarl/go-bindata-assetfs/"
	@GOPATH=$(GOPATH) go get -u "github.com/skelterjohn/geom"
	@GOPATH=$(GOPATH) go get -u "github.com/patrickmn/go-cache"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-http-mapzenjs"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-http-rewrite"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-geojson-v2"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-index"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-log"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-placetypes"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-spr"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-uri"
	rm -rf src/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/whosonfirst/go-whosonfirst-spr
	rm -rf src/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/whosonfirst/go-whosonfirst-flags
	rm -rf src/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/skelterjohn/geom
	rm -rf src/github.com/jteeuwen/go-bindata/testdata

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt app/*.go
	go fmt cache/*.go
	go fmt cmd/*.go
	go fmt filter/*.go
	go fmt flags/*.go
	go fmt http/*.go
	go fmt index/*.go
	go fmt utils/*.go
	go fmt *.go

bin: 	assets
	@GOPATH=$(GOPATH) go build -o bin/wof-pip cmd/wof-pip.go
	@GOPATH=$(GOPATH) go build -o bin/wof-pip-server cmd/wof-pip-server.go

assets:	self
	@GOPATH=$(GOPATH) go build -o bin/go-bindata ./vendor/github.com/jteeuwen/go-bindata/go-bindata/
	@GOPATH=$(GOPATH) go build -o bin/go-bindata-assetfs vendor/github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs/main.go
	rm -f www/*~ www/css/*~ www/javascript/*~
	@PATH=$(PATH):$(CWD)/bin bin/go-bindata-assetfs -pkg http www www/javascript www/css www/tangram
	mv bindata_assetfs.go http/assets.go

maps: wwwdirs mapzenjs tangram refill

wwwdirs:
	if test ! -d www/javascript; then mkdir -p www/javascript; fi
	if test ! -d www/css; then mkdir -p www/css; fi
	if test ! -d www/tangram; then mkdir -p www/tangram; fi

tangram:
	curl -s -o www/javascript/tangram.js https://mapzen.com/tangram/tangram.debug.js
	curl -s -o www/javascript/tangram.min.js https://mapzen.com/tangram/tangram.min.js

refill:
	curl -s -o www/tangram/refill-style.zip https://mapzen.com/carto/refill-style/refill-style.zip

mapzenjs:
	curl -s -o www/css/mapzen.js.css https://mapzen.com/js/mapzen.css
	curl -s -o www/javascript/mapzen.js https://mapzen.com/js/mapzen.js
	curl -s -o www/javascript/mapzen.min.js https://mapzen.com/js/mapzen.min.js
