# Connbeat

[![Build Status](https://travis-ci.org/raboof/connbeat.svg?branch=master)](https://travis-ci.org/raboof/connbeat)

Connbeat, short for 'Connectionbeat', is an open source agent that monitors connection metadata and
ships the data to Kafka or Elasticsearch, or a HTTP endpoint.

The main distinction from [Packetbeat](https://www.elastic.co/products/beats/packetbeat)
is that Connbeat is intended to be able to monitor all connections on a
machine (rather than just selected protocols), and does not inspect the
package/connection contents, only metadata.

## Status

The software is functional, but battle-testing and performance tuning is still in progress.

## Building

You need at least golang 1.7.3.

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

## Performance overhead

We tested the overhead of running the connbeat agent using the
[TechEmpower web framework benchmarks](https://www.techempower.com/benchmarks/).

After deploying to AWS, we ran the [query](https://www.techempower.com/benchmarks/#test=query)
benchmark workload against the Spring Boot framework.

The result was encouraging: the total requests throughput took a hit of only
0.47% (58 fewer requests on a total of 12312). The average latency was in fact
a little better in the test runs with connbeat - which must of course be caused
by noise, but inspires confidence that connbeat introduce no noticable degredation.

The complete test results can be found in the /tests/performance folder of this repo.

Of course performance impact may vary due to all kinds of circumstances and
differences in workload. We're aware of several potential further
optimizations, which can be applied when a situation comes up where connbeat
does have a noticable impact. If you encounter such a situation, be sure to
file an issue.

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

## Packaging

Preliminary packaging is available, but the resulting packages are not yet
intended for general consumption.

'make package' should be sufficient to produce a deb, rpm and a binary .tar.gz

## Contributing

Contributions are welcome! Feel free to [submit issues](https://github.com/raboof/connbeat/issues) to discuss problems and propose solutions, or send a [pull request](https://github.com/raboof/connbeat/pulls).

Pull requests are expected to include tests (which are run on Travis). We strive to merge any reasonable features, though features that might increase the load on the machine will likely have to be behind a feature switch that is off by default.
>>>>>>> origin/master
