# warning [![GoDoc](https://godoc.org/github.com/lunemec/warning?status.svg)](http://godoc.org/github.com/lunemec/warning) [![Go Report Card](https://goreportcard.com/badge/github.com/lunemec/warning)](https://goreportcard.com/report/github.com/lunemec/warning)
Package warning provides a simple way to handle errors that should not stop
execution (return err), but rather continue.

`go get github.com/lunemec/warning`

Common Go idiom is this:
```go
if err != nil {
    return err
}
```

But what if you wanted to distinguish between error that ends execution and
error that should just be logged?
This package provides you with just that.

```go
if err != nil && !warning.IsWarning(err) {
	//This is executed only if err is not a warning.
	return
}
```

It also works well with https://github.com/pkg/errors and https://github.com/hashicorp/go-multierror.
