# go-whosonfirst-placetypes

Go package for working with Who's On First placetypes.

## Example

### Simple

```
import (
       "github.com/whosonfirst/go-whosonfirst-placetypes"
       "log"
)

log.Println(placetypes.IsValidPlacetype("county"))
log.Println(placetypes.IsValidPlacetype("microhood"))
log.Println(placetypes.IsValidPlacetype("accelerator"))

id := int64(102312307)
log.Println(placetypes.IsValidPlacetypeId(id))          
```

Yields:

```
true
true
false
true
```

## See also

* https://github.com/whosonfirst/whosonfirst-placetypes
* https://github.com/whosonfirst/py-mapzen-whosonfirst-placetypes
