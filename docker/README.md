# Monitoring docker instances

connbeat can monitor connections in docker instances from the docker host or
from a 'peer' container (see below).

## Limitations

connbeat currently only supports docker 'bridge' networking. When monitoring
a docker instance you will collect only container metadata, no metadata on
individual processes within the container.

## Configuration

To enable monitoring docker instances set `enable_docker` to `true`. To avoid
also monitoring the non-docker connections you can set
`enable_local_connections` to `false`.

connbeat will report the hostname and IP address of the docker host (the host
on which docker itself runs) as exposed by 'docker info'. If this does not show
the full name (for example when running docker in the 'moby' VM on OSX) this
value can be overridden with the `DOCKERHOST_HOSTNAME` and `DOCKERHOST_IP`
environment variables.

If you want to expose any of the environment variables that have been passed to
the docker instances that are being monitored, add a whitelist to the
`docker_environment` configuration option.

## output

The connbeat output is the same as when monitoring regular connections, but
adds an additional section with container metadata:

```
"container": {
  "docker_host": {
    "hostname": "yinka",
    "ips": [
      "127.0.1.1"
    ]
  },
  "env": [
    "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
  ],
  "id": "0ca2e7481f6e230b1913ea8c08a2fb5481e5afc1101bbb4e52b1201ad7edb818",
  "image": "ubuntu",
  "local_ips": [
    "172.17.0.2"
  ],
  "names": [
    "/reverent_sinoussi"
  ],
  "ports": [
    {
      "80": [
        {
          "HostIp": "0.0.0.0",
          "HostPort": "1234"
        }
      ]
    }
  ]
```

As you can see the container id, image, names and portmappings are reported. If
more docker details are required it might be worth checking out
[dockbeat](https://github.com/Ingensi/dockbeat)

## docker peer monitoring

The 'linux' release of connbeat can be run inside this docker container to
monitor any peer containers. You need to mount the docker socket and provide
the target endpoint as an environment variable:

    docker run
      --rm
      -v /var/run/docker.sock:/var/run/docker.sock
      -e CONNBEAT_URL=pi.bzzt.net:80/foo
      raboof/connbeat:latest

### Building

Make sure a statically linked version of 'connbeat' is available to include in
the image, for example from the linux package obtained with 'make package'

Then simply build the docker image with:

    make

To do this in one go use 'make docker_peers' in the parent dir

### Deploying

After building, you can:

    docker push raboof/connbeat
