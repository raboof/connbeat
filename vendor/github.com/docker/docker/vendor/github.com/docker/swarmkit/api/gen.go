package api

//go:generate protoc -I.:../protobuf:../vendor:../vendor/github.com/gogo/protobuf --gogoswarm_out=plugins=grpc+deepcopy+storeobject+raftproxy+authenticatedwrapper,import_path=github.com/docker/swarmkit/api,Mgogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor,Mplugin/plugin.proto=github.com/docker/swarmkit/protobuf/plugin,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types:. types.proto specs.proto objects.proto control.proto dispatcher.proto ca.proto snapshot.proto raft.proto health.proto resource.proto logbroker.proto
