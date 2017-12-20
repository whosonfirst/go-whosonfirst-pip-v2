# https://blog.docker.com/2016/09/docker-golang/
# https://blog.golang.org/docker

# docker build -t wof-pip-server .

# $> docker run -p 6161:8080 -e HOST='0.0.0.0' -e EXTRAS='allow' -e MODE='sqlite' -e SOURCES='microhood-20171212' wof-pip-server
#
# fetch https://whosonfirst.mapzen.com/sqlite/microhood-20171212.db
# /go-whosonfirst-pip-v2/bin/wof-pip-server -host 0.0.0.0 -allow-extras -mode sqlite /usr/local/data/microhood-20171212.db
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

FROM golang

ADD . /go-whosonfirst-pip-v2

RUN cd /go-whosonfirst-pip-v2; make bin

RUN apt-get update && apt-get dist-upgrade -y && apt-get install -y bzip2

VOLUME /usr/local/data

EXPOSE 8080

ENTRYPOINT /go-whosonfirst-pip-v2/docker/entrypoint.sh

