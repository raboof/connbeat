# Connectionbeat

[![Build Status](https://travis-ci.org/raboof/connbeat.svg?branch=master)](https://travis-ci.org/raboof/connbeat)

Connectionbeat is an open source agent that monitors connection metadata and
ships the data to Kafka or Elasticsearch, or a HTTP endpoint.

The main distinction from [Packetbeat](https://www.elastic.co/products/beats/packetbeat)
is that Connectionbeat is intended to be able to monitor all connections on a
machine (rather than just selected protocols), and does not inspect the
package/connection contents, only metadata.

## Status

The software is functional, but battle-testing and performance tuning is still in progress.

## Building

You need at least golang 1.7 (see also: http://stackoverflow.com/questions/38922080/how-can-i-fallback-to-a-go-implementation-when-cgo-is-not-available-during-build)

    # Make sure $GOPATH is set
    go get github.com/raboof/connbeat
    cd $GOPATH/src/github.com/raboof/connbeat
    go get -t $(glide novendor)
    make

## Running

Edit the configuration (connbeat.yml) to specify where you want your events to go (e.g. Kafka, Elasticsearch, the console).

You need to be root if you want to see the process for processes other than your own:

    sudo ./connbeat

You can view the events on kafka with something like kafkacat:

    kafkacat -C -b localhost -t connbeat

## Events

For connections where the agent is the server:

    {
      "@timestamp": "2016-05-20T14:54:29.442Z",
      "beat": {
        "hostname": "yinka",
        "name": "yinka",
        "local_ips": [
          "192.168.2.243"
        ]
      },
      "local_port": 80,
      "local_process": {
        "binary": "dnsmasq",
        "cmdline": ""/usr/sbin/dnsmasq -x /var/run/dnsmasq/dnsmasq.pid -u dnsmasq -7 /etc/dnsmasq.d,.dpkg-dist,.dpkg-old,.dpkg-new --local-service",
        "environ": [
        "LANGUAGE=en_US:en",
        "PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin",
        "LANG=en_US.UTF-8",
        "_SYSTEMCTL_SKIP_REDIRECT=true",
        "PWD=/",

        ]
      },
      "type": "connbeat"
    }

For connections where the agent appears to be the client:

    {
      "@timestamp": "2016-05-20T14:54:29.506Z",
      "beat": {
        "hostname": "yinka",
        "name": "yinka",
        "local_ips": [
          "192.168.2.243"
        ]
      },
      "local_ip": "192.168.2.243",
      "local_port": 40074,
      "local_process": {
        "binary": "chromium",
        "cmdline": "/usr/lib/chromium/chromium --show-component-extension-options --ignore-gpu-blacklist --ppapi-flash-path=/usr/lib/pepperflashplugin-nonfree/libpepflashplayer.so --ppapi-flash-version=20.0.0.228",
        "environ": [
          ""
        ]
      },
      "remote_ip": "52.91.150.74",
      "remote_port": 443,
      "type": "connbeat"
    }

## Testing

To run the regular go unit test, run 'make test'.

To also run docker-based system tests, run 'make testsuite'

## Contributing

Contributions are welcome! Feel free to [submit issues](https://github.com/raboof/connbeat/issues) to discuss problems and propose solutions, or send a [pull request](https://github.com/raboof/connbeat/pulls).

Pull requests are expected to include tests (which are run on Travis). We strive to merge any reasonable features, though features that might increase the load on the machine will likely have to be behind a feature switch that is off by default.
