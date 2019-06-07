fmt:
	go fmt app/*.go
	go fmt cache/*.go
	go fmt cmd/*.go
	go fmt extras/*.go
	go fmt filter/*.go
	go fmt flags/*.go
	go fmt http/*.go
	go fmt index/*.go
	go fmt utils/*.go
	go fmt *.go

bin:
	go build -mod vendor -o bin/wof-pip cmd/wof-pip/main.go
	go build -mod vendor -o bin/wof-pip-server cmd/wof-pip-server/main.go

assets:
	go build -o bin/go-bindata ./vendor/github.com/whosonfirst/go-bindata/go-bindata/
	go build -o bin/go-bindata-assetfs vendor/github.com/whosonfirst/go-bindata-assetfs/go-bindata-assetfs/main.go
	rm -f www/*~ www/css/*~ www/javascript/*~
	@PATH=$(PATH):$(CWD)/bin bin/go-bindata-assetfs -pkg http www www/javascript www/css www/tangram
	mv bindata.go http/assets.go

maps: wwwdirs mapzenjs tangram refill

wwwdirs:
	if test ! -d www/javascript; then mkdir -p www/javascript; fi
	if test ! -d www/css; then mkdir -p www/css; fi
	if test ! -d www/tangram; then mkdir -p www/tangram; fi

tangram:
	curl -s -o www/javascript/tangram.js https://www.nextzen.org/tangram/tangram.debug.js
	curl -s -o www/javascript/tangram.min.js https://www.nextzen.org/tangram/tangram.min.js

refill:
	curl -s -o www/tangram/refill-style.zip https://www.nextzen.org/carto/refill-style/refill-style.zip

mapzenjs:
	curl -s -o www/css/mapzen.js.css https://www.nextzen.org/js/mapzen.css
	curl -s -o www/javascript/mapzen.js https://www.nextzen.org/js/mapzen.js
	curl -s -o www/javascript/mapzen.min.js https://www.nextzen.org/js/mapzen.min.js

crosshairs:
	curl -s -o www/javascript/slippymap.crosshairs.js https://raw.githubusercontent.com/whosonfirst/js-slippymap-crosshairs/master/src/slippymap.crosshairs.js	
