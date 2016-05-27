BEATNAME=connbeat
BEAT_DIR=github.com/raboof
SYSTEM_TESTS=false

# Only crosscompile for linux because other OS'es use cgo.
GOX_OS=linux

include ../../elastic/beats/libbeat/scripts/Makefile

# This is called by the beats packer before building starts
.PHONY: before-build
before-build:

