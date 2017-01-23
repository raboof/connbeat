all: image

.PHONY: image
image: connbeat connbeat.yml Dockerfile
	docker build . -t raboof/connbeat:latest
