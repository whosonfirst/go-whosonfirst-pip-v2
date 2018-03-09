#!/bin/sh

PIP_SERVER="/bin/wof-pip-server"
PIP_ARGS=""

DATA="/usr/local/data"

CURL=`which curl`
BUNZIP2=`which bunzip2`

if [ "${WOF_MODE}" = "sqlite" ]
then

    for DB in $(echo ${SQLITE_DATABASES} | sed "s/,/ /g")
    do
	
	REMOTE="https://dist.whosonfirst.org/sqlite/${DB}"
	LOCAL="${DATA}/${DB}"
	
	if [ ! -f ${LOCAL} ]
	then
	    echo "fetch ${REMOTE}.bz2"
	    
	    ${CURL} -o ${LOCAL}.bz2 ${REMOTE}.bz2
	    
	    if [ $? -ne 0 ]
	    then
		echo "failed to fetch remote source ${REMOTE}.bz2"
		continue
	    fi
	    
	    ${BUNZIP2} ${LOCAL}.bz2
	    
	    if [ $? -ne 0 ]
	    then
		echo "failed to uncompress local source"
		exit 0
	    fi
	    
	fi
	
	PIP_ARGS="${PIP_ARGS} ${LOCAL}"    
    done

elif [ "${WOF_MODE}" = "spatialite" ]
then

    REMOTE="https://dist.whosonfirst.org/sqlite/${SPATIALITE_DATABASE}"
    LOCAL="${DATA}/${SPATIALITE_DATABASE}"
	
    if [ ! -f ${LOCAL} ]
    then
	echo "fetch ${REMOTE}.bz2 as ${LOCAL}.bz2"
	
	${CURL} -o ${LOCAL}.bz2 ${REMOTE}.bz2
	
	if [ $? -ne 0 ]
	then
	    echo "failed to fetch remote source ${REMOTE}.bz2"
	    exit 0
	fi
	
	${BUNZIP2} ${LOCAL}.bz2
	
	if [ $? -ne 0 ]
	then
	    echo "failed to uncompress local source"
	    exit 0
	fi
	
    fi
    
    export WOF_SPATIALITE_DSN="${LOCAL}"
    export LD_LIBRARY_PATH=".:/lib:/usr/lib:/usr/local/lib"
  
else
    echo "only '-mode sqlite' or '-mode spatialite' are supported right now"
    exit 1
fi

${PIP_SERVER} -setenv ${PIP_ARGS}

if [ $? -ne 0 ]
then
   echo "command '${PIP_SERVER} ${PIP_ARGS}' failed"
   exit 1
fi

exit 0