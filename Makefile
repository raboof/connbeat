BEATNAME=connbeat
BEAT_DIR=github.com/raboof
ES_BEATS=../../elastic/beats
SYSTEM_TESTS=true
TEST_ENVIRONMENT?=true
DOCKER_COMPOSE=docker-compose -f ../../elastic/beats/testing/environments/base.yml -f ../../elastic/beats/testing/environments/${TESTING_ENVIRONMENT}.yml -f docker-compose.yml


# Only crosscompile for linux because other OS'es use cgo.
GOX_OS=linux

include ../../elastic/beats/libbeat/scripts/Makefile

# This is called by the beats packer before building starts
.PHONY: before-build
before-build:
