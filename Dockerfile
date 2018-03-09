# https://blog.docker.com/2016/09/docker-golang/
# https://blog.golang.org/docker

# docker build -t wof-pip-server .

# $> docker run -it -p 6161:8080 -e WOF_HOST='0.0.0.0' -e WOF_ENABLE_EXTRAS='true' -e WOF_MODE='sqlite' -e SQLITE_DATABASES='microhood-20171212.db' wof-pip-server
#
# fetch https://whosonfirst.mapzen.com/sqlite/microhood-20171212.db
# /go-whosonfirst-pip-v2/bin/wof-pip-server -host 0.0.0.0 -enable-extras=true -mode sqlite /usr/local/data/microhood-20171212.db
# 23:05:42.065812 [wof-pip-server] STATUS create temporary extras database '/tmp/pip-extras558496578'
# 23:05:42.067233 [wof-pip-server] STATUS listening on 0.0.0.0:8080
# 23:05:43.068415 [wof-pip-server] STATUS indexing 385 records indexed
# 23:05:44.068214 [wof-pip-server] STATUS indexing 661 records indexed
# ...
# 23:05:49.068497 [wof-pip-server] STATUS indexing 1509 records indexed
# 23:05:49.687971 [wof-pip-server] STATUS finished indexing
#
# and then:
# $> curl 'http://localhost:6161/?latitude=37.794906&longitude=-122.395229&extras=name:,edtf:' | python -mjson.tool

# $> docker run -it -p 6161:8080 -e WOF_HOST='0.0.0.0' -e WOF_INDEX='spatialite' -e WOF_CACHE='spatialite' -e WOF_MODE='spatialite' -e SPATIALITE_DATABASE='whosonfirst-data-constituency-us-latest.db' wof-pip-server

# build phase - see also:
# https://medium.com/travis-on-docker/multi-stage-docker-builds-for-creating-tiny-go-images-e0e1867efe5a
# https://medium.com/travis-on-docker/triple-stage-docker-builds-with-go-and-angular-1b7d2006cb88

FROM golang:alpine AS build-env

# https://github.com/gliderlabs/docker-alpine/issues/24

RUN apk add --update alpine-sdk

ADD . /go-whosonfirst-pip-v2

RUN cd /go-whosonfirst-pip-v2; make bin

FROM alpine

RUN apk add --update bzip2 curl

# PLEASE HELP ME FIGURE OUT WHY THIS DOESN'T WORK (20180309/thisisaaronland)
# SPECIFICALLY BY THE TIME WE TRY TO LOAD libspatialite IT FAILS WITH THE FOLLOWING ERROR:
# Failed to create new PIP application, because shaxbee/go-spatialite: spatialite extension not found.
# WHICH IS TRIGGERED HERE:
# https://github.com/whosonfirst/go-whosonfirst-sqlite/blob/master/vendor/github.com/whosonfirst/go-spatialite/spatialite.go
#
# RUN apk add --update --repository http://dl-3.alpinelinux.org/alpine/edge/testing/ libspatialite

VOLUME /usr/local/data

WORKDIR /bin/

COPY --from=build-env /go-whosonfirst-pip-v2/bin/wof-pip-server /bin/wof-pip-server
COPY --from=build-env /go-whosonfirst-pip-v2/docker/entrypoint.sh /bin/entrypoint.sh

EXPOSE 8080

ENTRYPOINT /bin/entrypoint.sh

