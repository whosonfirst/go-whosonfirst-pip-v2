# https://blog.docker.com/2016/09/docker-golang/
# https://blog.golang.org/docker

# docker build -t wof-pip .

FROM golang

ADD . /go-whosonfirst-pip-v2

RUN cd /go-whosonfirst-pip-v2; make bin

EXPOSE 8080

ENTRYPOINT /go-whosonfirst-pip-v2/docker/entrypoint.sh

