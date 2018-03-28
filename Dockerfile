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

RUN apk add --update make libc-dev gcc

ADD . /go-whosonfirst-pip-v2
RUN cd /go-whosonfirst-pip-v2; make bin

# https://github.com/terranodo/spatialite-docker/blob/master/Dockerfile

RUN echo "@edge http://nl.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories
RUN echo "@edge-testing http://nl.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk update

RUN apk add wget gcc make libc-dev sqlite-dev zlib-dev libxml2-dev "proj4-dev@edge-testing" "geos-dev@edge-testing" "gdal-dev@edge-testing" "gdal@edge-testing" expat-dev readline-dev ncurses-dev ncurses-static libc6-compat

RUN wget "http://www.gaia-gis.it/gaia-sins/freexl-sources/freexl-1.0.5.tar.gz" && tar zxvf freexl-1.0.5.tar.gz && cd freexl-1.0.5 && ./configure && make && make install

RUN wget "http://www.gaia-gis.it/gaia-sins/libspatialite-4.3.0a.tar.gz" && tar zxvf libspatialite-4.3.0a.tar.gz && cd libspatialite-4.3.0a && ./configure && make && make install

RUN wget "http://www.gaia-gis.it/gaia-sins/readosm-1.1.0.tar.gz" && tar zxvf readosm-1.1.0.tar.gz && cd readosm-1.1.0 && ./configure && make && make install

RUN wget "http://www.gaia-gis.it/gaia-sins/spatialite-tools-4.3.0.tar.gz" && tar zxvf spatialite-tools-4.3.0.tar.gz && cd spatialite-tools-4.3.0 && ./configure && make && make install

RUN cp /usr/local/bin/* /usr/bin/
RUN cp -R /usr/local/lib/* /usr/lib/

FROM alpine

RUN apk add --update bzip2 curl

VOLUME /usr/local/data

WORKDIR /bin/

COPY --from=build-env /usr/lib/ /usr/lib
COPY --from=build-env /usr/bin/ /usr/bin

COPY --from=build-env /go-whosonfirst-pip-v2/bin/wof-pip-server /bin/wof-pip-server
COPY --from=build-env /go-whosonfirst-pip-v2/docker/entrypoint.sh /bin/entrypoint.sh

EXPOSE 8080

ENTRYPOINT /bin/entrypoint.sh

