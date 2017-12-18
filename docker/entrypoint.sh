#!/bin/sh

# This assumes 'ADD . /go-whosonfirst-pip-v2' which is defined in the Dockerfile

PIP_SERVER="/go-whosonfirst-pip-v2/bin/wof-pip-server"
ARGS=""

if [ "${HOST}" != "" ]
then
    ARGS="${ARGS} -host ${HOST}"
fi

if [ "${WWW}" != "" ]
then
    ARGS="${ARGS} -www"

    if [ "${MAPZEN_APIKEY}" != "" ]
    then
	ARGS="${ARGS} -mapzen-apikey ${MAPZEN_APIKEY}"
    fi
    
fi

${PIP_SERVER} ${ARGS} -mode sqlite /go-whosonfirst-pip-v2/data/*.db

if [ $? -ne 0 ]
then
   echo "command '${STATICD} ${ARGS}' failed"
   exit 1
fi

exit 0

