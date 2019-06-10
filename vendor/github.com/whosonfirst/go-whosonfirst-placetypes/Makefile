vendor-deps: 
	go mod vendor

fmt:
	go fmt *.go
	go fmt placetypes/*.go
	go fmt filter/*.go

tools:	
	go build -o bin/wof-placetype-ancestors cmd/wof-placetype-ancestors/main.go
	go build -o bin/wof-placetype-children cmd/wof-placetype-children/main.go
	go build -o bin/wof-placetype-descendants cmd/wof-placetype-descendants/main.go

spec:
	go run cmd/mk-spec.main.go > placetypes/spec.go
