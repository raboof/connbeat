BEATNAME=connbeat
BEAT_DIR=github.com/raboof/connbeat
SYSTEM_TESTS=true
TEST_ENVIRONMENT?=true
ES_BEATS?=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell go list ${BEAT_DIR}/... 2>/dev/null | grep -v /vendor/)
PREFIX?=.

# Only crosscompile for linux because other OS'es use cgo.
GOX_OS=linux darwin
GOX_FLAGS='-arch=amd64 386'

# For packaging: for now we know how to package on linux amd64
TARGETS="linux/amd64 linux/386"
PACKAGES=connbeat/deb connbeat/rpm connbeat/bin

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

VERSION=$(shell ./vendor/github.com/elastic/beats/dev-tools/get_version | sed -e s/-.*//)-$(shell git rev-parse --short HEAD)

update: my_update_extension

my_update_extension:
	cat _meta/beat.head.yml vendor/github.com/raboof/beats-output-http/configuration_example.yml > _meta/beat.yml

.PHONY: update_version
update_version:
	./vendor/github.com/elastic/beats/dev-tools/set_version ${VERSION}
	sed -i '/"name": "SNAPSHOT",/c\"name": "${VERSION}",' descriptor.bintray

.PHONY: docker_peers
docker_peers: package
	@rm -r unpacked || true
	mkdir unpacked
	tar xzvf build/upload/connbeat-*-linux-x86.tar.gz -C unpacked
	cp unpacked/*/connbeat docker
	$(MAKE) -C docker
