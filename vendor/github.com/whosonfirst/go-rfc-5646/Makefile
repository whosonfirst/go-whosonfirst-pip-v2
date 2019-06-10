vendor-deps: 
	go mod vendor

fmt:
	go fmt *.go
	go fmt cmd/*.go
	go fmt subtags/*.go
	go fmt tags/*.go

tools: 	
	go build -o bin/rfc-5646-parse cmd/rfc-5646-parse.go
