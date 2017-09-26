# go-http-rewrite

Go middleware package for HTTP rewrite rules.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Example

_Please write me._

## Interfaces

### RewriteRule

```
type RewriteRule interface {
     Match(string) bool
     Transform(string) string
     Last() bool
}
```
