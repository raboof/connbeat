BEATNAME=connbeat
BEAT_DIR=github.com/raboof/connbeat
SYSTEM_TESTS=true
TEST_ENVIRONMENT?=true
ES_BEATS?=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell go list ${BEAT_DIR}/... 2>/dev/null | grep -v /vendor/)
DOCKER_COMPOSE=docker-compose -f vendor/github.com/elastic/beats/testing/environments/base.yml -f vendor/github.com/elastic/beats/testing/environments/${TESTING_ENVIRONMENT}.yml -f docker-compose.yml
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

# This is called by the beats packer before starts
.PHONY: build before-build
before-build:

# Collects all dependencies and then calls update
.PHONY: collect
collect:

