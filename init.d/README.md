# init.d

## wof-pip-proxy-start.sh

Here's an example set of configs

```
# Stuff you will need to edit

WOF_DATA=/usr/local/data/whosonfirst-data/data

PIP_DAEMON=/usr/local/mapzen/go-whosonfirst-pip/bin/wof-pip-server

PROXY_DAEMON=/usr/local/mapzen/go-whosonfirst-pip/bin/wof-pip-proxy
PROXY_CONFIG=/usr/local/mapzen/config/wof-pip-proxy.json

START_USER=www-data
START_DAEMON=/usr/local/mapzen/go-whosonfirst-pip/utils/wof-pip-proxy-start.py

START_ARGS="" 

# Okay - you shouldn't need to edit anything after this
```

Two other things to keep in mind.

First you will need to create a config file using `mk-wof-config.py`, like this:

```
./utils/mk-wof-config.py -w /usr/local/data/whosonfirst-data -o - -r common,common_optional > /usr/local/mapzen/config/wof-pip-proxy.json
```

Second make sure that all the meta files (not to mention the GeoJSON files) are readable by the user running the proxy service. This should be obvious but as of this writing (20160324) some of the libraries we are using to write files atomically are insane about permissions and we're still sorting out that mess...