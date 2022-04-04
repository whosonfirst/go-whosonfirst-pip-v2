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

## Tools

```
$> make cli
go build -mod vendor -o bin/wof-placetype-ancestors cmd/wof-placetype-ancestors/main.go
go build -mod vendor -o bin/wof-placetype-children cmd/wof-placetype-children/main.go
go build -mod vendor -o bin/wof-placetype-descendants cmd/wof-placetype-descendants/main.go
go build -mod vendor -o bin/wof-valid-placetype cmd/wof-valid-placetype/main.go
```

### wof-placetype-ancestors

```
./bin/wof-placetype-ancestors -role common -role common_optional -role optional locality
2018/08/23 17:09:53 0 localadmin
2018/08/23 17:09:53 1 county
2018/08/23 17:09:53 2 macrocounty
2018/08/23 17:09:53 3 region
2018/08/23 17:09:53 4 macroregion
2018/08/23 17:09:53 5 dependency
2018/08/23 17:09:53 6 empire
2018/08/23 17:09:53 7 country
2018/08/23 17:09:53 8 continent
2018/08/23 17:09:53 9 disputed
2018/08/23 17:09:53 10 planet

./bin/wof-placetype-ancestors -role common -role common_optional -role optional enclosure
2018/08/23 17:10:42 0 venue
2018/08/23 17:10:42 1 arcade
2018/08/23 17:10:42 2 concourse
2018/08/23 17:10:42 3 wing
2018/08/23 17:10:42 4 building
2018/08/23 17:10:42 5 address
2018/08/23 17:10:42 6 intersection
2018/08/23 17:10:42 7 campus
2018/08/23 17:10:42 8 microhood
2018/08/23 17:10:42 9 neighbourhood
2018/08/23 17:10:42 10 macrohood
2018/08/23 17:10:42 11 borough
2018/08/23 17:10:42 12 locality
2018/08/23 17:10:42 13 localadmin
2018/08/23 17:10:42 14 county
2018/08/23 17:10:42 15 macrocounty
2018/08/23 17:10:42 16 region
2018/08/23 17:10:42 17 macroregion
2018/08/23 17:10:42 18 dependency
2018/08/23 17:10:42 19 empire
2018/08/23 17:10:42 20 country
2018/08/23 17:10:42 21 continent
2018/08/23 17:10:42 22 disputed
2018/08/23 17:10:42 23 planet
```

### wof-placetype-children

```
./bin/wof-placetype-children locality
2018/08/24 18:16:53 0 borough
2018/08/24 18:16:53 1 postalcode
2018/08/24 18:16:53 2 macrohood
2018/08/24 18:16:53 3 neighbourhood
2018/08/24 18:16:53 4 campus
```

### wof-placetype-descendants

```
./bin/wof-placetype-descendants -role common -role common_optional -role optional country
2018/08/24 18:15:49 0 marinearea
2018/08/24 18:15:49 1 timezone
2018/08/24 18:15:49 2 disputed
2018/08/24 18:15:49 3 macroregion
2018/08/24 18:15:49 4 region
2018/08/24 18:15:49 5 macrocounty
2018/08/24 18:15:49 6 county
2018/08/24 18:15:49 7 localadmin
2018/08/24 18:15:49 8 locality
2018/08/24 18:15:49 9 postalcode
2018/08/24 18:15:49 10 campus
2018/08/24 18:15:49 11 borough
2018/08/24 18:15:49 12 macrohood
2018/08/24 18:15:49 13 neighbourhood
2018/08/24 18:15:49 14 microhood
2018/08/24 18:15:49 15 intersection
2018/08/24 18:15:49 16 address
2018/08/24 18:15:49 17 building
2018/08/24 18:15:49 18 venue
2018/08/24 18:15:49 19 wing
2018/08/24 18:15:49 20 concourse
2018/08/24 18:15:49 21 arcade
2018/08/24 18:15:49 22 installation
2018/08/24 18:15:49 23 enclosure
```

### wof-valid-placetype

```
$> ./bin/wof-valid-placetype bob custom building locality
2021/02/19 17:35:05 bob	false
2021/02/19 17:35:05 custom	true
2021/02/19 17:35:05 building	true
2021/02/19 17:35:05 locality	true
```

## See also

* https://github.com/whosonfirst/whosonfirst-placetypes
* https://github.com/whosonfirst/py-mapzen-whosonfirst-placetypes
