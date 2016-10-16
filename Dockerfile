FROM golang:1.7.1

RUN set -x && \
    apt-get update && \
    apt-get install -y netcat python-virtualenv python-pip && \
    apt-get clean

## Install go package dependencies
RUN set -x \
  go get \
	github.com/pierrre/gotestcover \
	github.com/tsg/goautotest \
	golang.org/x/tools/cmd/vet \
 	github.com/Masterminds/glide

# Setup work environment
ENV CONNBEAT_PATH /go/src/github.com/raboof/connbeat
ENV GO15VENDOREXPERIMENT=1

RUN mkdir -p $CONNBEAT_PATH/build/coverage
WORKDIR $CONNBEAT_PATH
