# Connectionbeat

[![Build Status](https://travis-ci.org/raboof/connbeat.svg?branch=master)](https://travis-ci.org/raboof/connbeat)

Connectionbeat is an open source agent that monitors connection metadata and
ships the data to Kafka or Elasticsearch.

The main distinction from [Packetbeat](https://www.elastic.co/products/beats/packetbeat)
is that Connectionbeat is intended to be able to monitor all connections on a
machine (rather than just selected protocols), and does not inspect the
package/connection contents, only metadata.

## Status

This is a proof-of-concept. While functional, battle-testing and performance tuning is still in progress.

## Building

    # Make sure $GOPATH is set
    go get github.com/raboof/connbeat
    cd $GOPATH/src/github.com/raboof/connbeat
    go get -t ./...
    make

## Running

The default configuration (connbeat.yml) logs to kafka on localhost:9092 and to the console.

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
