# go-whosonfirst-names

Go package for working with names and RFC 5646 language tags in Who's On First documents.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

Out of the box Who's On First name language tags are not RFC 5646 (sometimes called BCP 47) compliant.

This package follows the [RFC 5646 (BCP 47) comformance](https://github.com/whosonfirst/whosonfirst-names/#rfc-5646-bcp-47-comformance) and [RFC 5646 (BCP 47) tag conversion](https://github.com/whosonfirst/whosonfirst-names/#rfc-5646-bcp-47-tag-conversion) rules defined in the `whosonfirst-name` repository for transiting between the two.

As of this writing that does _not_ account for the fact that Who's On First uses private use (`-x-SOMETHING`) subtags that may be longer than the RFC 5646 limit of eight characters.

## Example

```
package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-names/tags"
	"github.com/whosonfirst/go-whosonfirst-names/utils"	
	"log"
)

func main() {

	flag.Parse()

	for _, raw := range flag.Args() {

		langtag, _ := tags.NewLangTag(raw)

		log.Println(langtag.String())
		log.Println(utils.ToRFC5646(langtag.String()))
	}
}
```

_Error handling removed for the sake of brevity._

## See also

* https://github.com/whosonfirst/whosonfirst-names
* https://github.com/whosonfirst/go-rfc-5646
