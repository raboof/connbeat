# docker peer monitoring

The 'linux' release of connbeat can be run inside this docker container to
monitor any peer containers. You need to mount the docker socket and provide
the target endpoint as an environment variable:

    docker run
      --rm
      -v /var/run/docker.sock:/var/run/docker.sock
      -e CONNBEAT_URL=pi.bzzt.net:80/foo
      raboof/connbeat:latest

## Building

Make sure a statically linked version of 'connbeat' is available to include in
the image, for example from the linux package obtained with 'make package'

Then simply build the docker image with:

    make

To do this in one go use 'make docker_peers' in the parent dir

## Deploying

After building, you can:

    docker push raboof/connbeat
