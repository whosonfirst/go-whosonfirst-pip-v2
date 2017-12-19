#!/bin/sh

# This assumes 'ADD . /go-whosonfirst-pip-v2' which is defined in the Dockerfile

PIP_SERVER="/go-whosonfirst-pip-v2/bin/wof-pip-server"
ARGS=""

CURL=`which curl`
BUNZIP2=`which bunzip2`

if [ "${HOST}" != "" ]
then
    ARGS="${ARGS} -host ${HOST}"
fi

if [ "${EXTRAS}" != "" ]
then
    ARGS="${ARGS} -extras"
fi

if [ "${WWW}" != "" ]
then
    ARGS="${ARGS} -www"

    if [ "${MAPZEN_APIKEY}" != "" ]
    then
	ARGS="${ARGS} -mapzen-apikey ${MAPZEN_APIKEY}"
    fi
    
fi

if [ "${MODE}" != "" ]
then
    ARGS="${ARGS} -mode ${MODE}"
fi

for DB in $(echo ${SOURCES} | sed "s/,/ /g")
do
    REMOTE="https://whosonfirst.mapzen.com/sqlite/${DB}.db"
    LOCAL="/usr/local/data/${DB}.db"

    if [ ! -f ${LOCAL} ]
    then
	echo "fetch ${REMOTE}"
    
	${CURL} -s -o ${LOCAL}.bz2 ${REMOTE}.bz2

	if [ $? -ne 0 ]
	then
	    echo "failed to fetch remote source"
	    exit 0
	fi
	
	${BUNZIP2} ${LOCAL}.bz2

	if [ $? -ne 0 ]
	then
	   echo "failed to uncompress local source"
	   exit 0
	fi
	   
    fi

    ARGS="${ARGS} ${LOCAL}"    
done

echo ${PIP_SERVER} ${ARGS}
${PIP_SERVER} ${ARGS}

if [ $? -ne 0 ]
then
   echo "command '${PIP_SERVER} ${ARGS}' failed"
   exit 1
fi

exit 0
