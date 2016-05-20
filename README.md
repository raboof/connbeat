# Connectionbeat

Connectionbeat is an open source agent that monitors connection metadata and
ships the data to Kafka or Elasticsearch.

The main distinction from [Packetbeat](https://www.elastic.co/products/beats/packetbeat)
is that Connectionbeat is intended to be able to monitor all connections on a
machine (rather than just selected protocols), and does not inspect the
package/connection contents, only metadata.

## Status

This is currently a PoC sketch, not in fact functional.
