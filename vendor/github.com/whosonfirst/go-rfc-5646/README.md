# go-rfc-5646

Go package for working with RFC 5646 language tags.

## Install

You will need to have both `Go` (specifically [version 1.12](https://golang.org/dl/) or higher because we're using [Go modules](https://github.com/golang/go/wiki/Modules)) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

This is a pretty naive implementation. It works for the common language tags, including extensions and private use subtags, but might not work for grandfathered or irregular tags.

## Example

```
package main

import (
	"flag"
	"github.com/whosonfirst/go-rfc-5646/tags"
	"log"
)

func main() {

	flag.Parse()

	for _, raw := range flag.Args() {

		langtag, _ := tags.NewLangTag(raw)

		log.Println(langtag.String())
		log.Println(langtag.Language())		
	}
}
```

_Note that error handling has been left out for the sake of brevity._

## Interfaces

As of this writing there is only one interface for language tags themselves, and all its methods return strings. Eventually there will be interfaces for the subtags that make up a language tag and the current interface will be updated accordingly.

### rfc5646.LanguageTag

```
type LanguageTag interface {
	Language() string
	ExtLang() string
	Script() string
	Region() string
	Variant() string
	Extension() string
	PrivateUse() string
	String() string
}
```

## See also

* https://www.rfc-editor.org/rfc/rfc5646.txt
* https://www.w3.org/International/articles/language-tags/
