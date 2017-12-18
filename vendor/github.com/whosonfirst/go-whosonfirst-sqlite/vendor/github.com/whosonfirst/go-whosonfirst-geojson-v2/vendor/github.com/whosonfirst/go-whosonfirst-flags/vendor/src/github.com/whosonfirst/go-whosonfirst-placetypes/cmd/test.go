package main

import (
       "github.com/whosonfirst/go-whosonfirst-placetypes"
       "log"
)

func main() {

     log.Println(placetypes.IsValidPlacetype("county"))
     log.Println(placetypes.IsValidPlacetype("microhood"))
     log.Println(placetypes.IsValidPlacetype("accelerator"))

     id := int64(102312307)
     log.Println(id)     	
     log.Println(placetypes.IsValidPlacetypeId(id))          
}
