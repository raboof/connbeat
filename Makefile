BEATNAME=connbeat
BEAT_DIR=github.com/raboof/connbeat
SYSTEM_TESTS=true
TEST_ENVIRONMENT?=true
ES_BEATS?=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
DOCKER_COMPOSE=docker-compose -f ../../elastic/beats/testing/environments/base.yml -f ../../elastic/beats/testing/environments/${TESTING_ENVIRONMENT}.yml -f docker-compose.yml
CGO=true
PREFIX?=.

# Only crosscompile for linux because other OS'es use cgo.
GOX_OS=linux

# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile

# Initial beat setup
.PHONY: setup
setup: copy-vendor
	make update

# Copy beats into vendor directory
.PHONY: copy-vendor
copy-vendor:
	mkdir -p vendor/github.com/elastic/
	cp -R ${GOPATH}/src/github.com/elastic/beats vendor/github.com/elastic/
	rm -rf vendor/github.com/elastic/beats/.git

# This is called by the beats packer before building starts
.PHONY: build before-build
before-build:

connbeat:
	go build -ldflags "-linkmode external -extldflags -static"
