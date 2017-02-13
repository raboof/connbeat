# Current overhead

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

# Potential improvements

## Nice/rtprio

We could set the nice/rtprio to make sure the 'real' workload is scheduled
before connbeat.

## Polling intervals

The polling intervals can be manipulated through the configuration. Polling
less frequently will reduce impact.

## Core usage

We could restrict the amount of cores connbeat is allowed to run on by setting
GOMAXPROCS

## Proc scanning

When enriching the connection information with process information is enabled,
a significant portion of the CPU usage is spent scanning /proc to find the
process to be associated with a given inode.

When not running as root, we won't have access to the details of processes
owned by other users. In that case we could consider caching the inodes and
PIDs to avoid checking them on every cache miss.

When the monitored system is for example a HTTP server, you could expect a lot
of requests to port 80 to be associated with the same process. We could collect
statistics on this and use it to first check that process instead of scanning
all of /proc each time.

## Cache eviction strategies

We could tune caches to limit the amount of elements they are allowed to
contain, reducing the memory pressure and the time it takes to check the cache.

