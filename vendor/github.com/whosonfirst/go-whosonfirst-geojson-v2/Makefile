vendor-deps: 
	go mod vendor

fmt:
	go fmt cmd/*.go
	go fmt feature/*.go
	go fmt geometry/*.go
	go fmt properties/geometry/*.go
	go fmt properties/whosonfirst/*.go
	go fmt utils/*.go
	go fmt *.go

tools:
	go build -o bin/wof-feature-to-spr cmd/wof-feature-to-spr/main.go
	go build -o bin/wof-geojson-dump cmd/wof-geojson-dump/main.go
	go build -o bin/wof-geojson-existential cmd/wof-geojson-existential/main.go
	go build -o bin/wof-geojson-hash cmd/wof-geojson-hash/main.go
	go build -o bin/wof-geojson-intersects cmd/wof-geojson-intersects/main.go
	go build -o bin/wof-geojson-names cmd/wof-geojson-names/main.go
