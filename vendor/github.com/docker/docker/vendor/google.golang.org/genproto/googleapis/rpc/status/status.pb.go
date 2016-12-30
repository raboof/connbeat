// Code generated by protoc-gen-go.
// source: google.golang.org/genproto/googleapis/rpc/status/status.proto
// DO NOT EDIT!

/*
Package google_rpc is a generated protocol buffer package.

It is generated from these files:
	google.golang.org/genproto/googleapis/rpc/status/status.proto

It has these top-level messages:
	Status
*/
package google_rpc // import "google.golang.org/genproto/googleapis/rpc/status"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// The `Status` type defines a logical error model that is suitable for different
// programming environments, including REST APIs and RPC APIs. It is used by
// [gRPC](https://github.com/grpc). The error model is designed to be:
//
// - Simple to use and understand for most users
// - Flexible enough to meet unexpected needs
//
// # Overview
//
// The `Status` message contains three pieces of data: error code, error message,
// and error details. The error code should be an enum value of
// [google.rpc.Code][google.rpc.Code], but it may accept additional error codes if needed.  The
// error message should be a developer-facing English message that helps
// developers *understand* and *resolve* the error. If a localized user-facing
// error message is needed, put the localized message in the error details or
// localize it in the client. The optional error details may contain arbitrary
// information about the error. There is a predefined set of error detail types
// in the package `google.rpc` which can be used for common error conditions.
//
// # Language mapping
//
// The `Status` message is the logical representation of the error model, but it
// is not necessarily the actual wire format. When the `Status` message is
// exposed in different client libraries and different wire protocols, it can be
// mapped differently. For example, it will likely be mapped to some exceptions
// in Java, but more likely mapped to some error codes in C.
//
// # Other uses
//
// The error model and the `Status` message can be used in a variety of
// environments, either with or without APIs, to provide a
// consistent developer experience across different environments.
//
// Example uses of this error model include:
//
// - Partial errors. If a service needs to return partial errors to the client,
//     it may embed the `Status` in the normal response to indicate the partial
//     errors.
//
// - Workflow errors. A typical workflow has multiple steps. Each step may
//     have a `Status` message for error reporting purpose.
//
// - Batch operations. If a client uses batch request and batch response, the
//     `Status` message should be used directly inside batch response, one for
//     each error sub-response.
//
// - Asynchronous operations. If an API call embeds asynchronous operation
//     results in its response, the status of those operations should be
//     represented directly using the `Status` message.
//
// - Logging. If some API errors are stored in logs, the message `Status` could
//     be used directly after any stripping needed for security/privacy reasons.
type Status struct {
	// The status code, which should be an enum value of [google.rpc.Code][google.rpc.Code].
	Code int32 `protobuf:"varint,1,opt,name=code" json:"code,omitempty"`
	// A developer-facing error message, which should be in English. Any
	// user-facing error message should be localized and sent in the
	// [google.rpc.Status.details][google.rpc.Status.details] field, or localized by the client.
	Message string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
	// A list of messages that carry the error details.  There will be a
	// common set of message types for APIs to use.
	Details []*google_protobuf.Any `protobuf:"bytes,3,rep,name=details" json:"details,omitempty"`
}

func (m *Status) Reset()                    { *m = Status{} }
func (m *Status) String() string            { return proto.CompactTextString(m) }
func (*Status) ProtoMessage()               {}
func (*Status) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Status) GetDetails() []*google_protobuf.Any {
	if m != nil {
		return m.Details
	}
	return nil
}

func init() {
	proto.RegisterType((*Status)(nil), "google.rpc.Status")
}

func init() {
	proto.RegisterFile("google.golang.org/genproto/googleapis/rpc/status/status.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 208 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x44, 0x8e, 0xcd, 0x4e, 0x84, 0x30,
	0x10, 0xc7, 0x83, 0xab, 0xbb, 0xb1, 0x9b, 0x78, 0x68, 0x3c, 0x34, 0x5e, 0xdc, 0x78, 0xe2, 0x34,
	0x93, 0xe8, 0xd9, 0x83, 0x3c, 0x01, 0xc1, 0x27, 0x28, 0x50, 0x46, 0x12, 0xe8, 0x34, 0xb4, 0x1c,
	0x78, 0x7b, 0xa1, 0x85, 0xec, 0xa1, 0x69, 0x3b, 0xf3, 0xfb, 0x7f, 0x88, 0x6f, 0x62, 0xa6, 0xc1,
	0x00, 0xf1, 0xa0, 0x2d, 0x01, 0x4f, 0x84, 0x64, 0xac, 0x9b, 0x38, 0x30, 0xa6, 0x95, 0x76, 0xbd,
	0xc7, 0xc9, 0x35, 0xe8, 0x83, 0x0e, 0xb3, 0xdf, 0x2f, 0x88, 0x88, 0x14, 0xbb, 0x7c, 0xdd, 0xbf,
	0x21, 0xf5, 0xe1, 0x6f, 0xae, 0xa1, 0xe1, 0x11, 0x93, 0x1d, 0x46, 0xa8, 0x9e, 0x3b, 0x74, 0x61,
	0x71, 0xc6, 0xa3, 0xb6, 0xcb, 0x76, 0x92, 0xf8, 0xa3, 0x13, 0xe7, 0xdf, 0x68, 0x26, 0xa5, 0x78,
	0x6c, 0xb8, 0x35, 0x2a, 0xbb, 0x65, 0xf9, 0x53, 0x15, 0xdf, 0x52, 0x89, 0xcb, 0x68, 0xbc, 0xd7,
	0x64, 0xd4, 0xc3, 0x3a, 0x7e, 0xae, 0x8e, 0xaf, 0x04, 0x71, 0x69, 0x4d, 0xd0, 0xfd, 0xe0, 0xd5,
	0xe9, 0x76, 0xca, 0xaf, 0x9f, 0xaf, 0xb0, 0xd7, 0x38, 0xf2, 0xe0, 0xc7, 0x2e, 0xd5, 0x01, 0x15,
	0xef, 0xe2, 0x65, 0xed, 0x04, 0xf7, 0xaa, 0xc5, 0x35, 0xe5, 0x96, 0x1b, 0x5e, 0x66, 0xf5, 0x39,
	0xea, 0xbe, 0xfe, 0x03, 0x00, 0x00, 0xff, 0xff, 0x73, 0x63, 0xb7, 0xba, 0x0d, 0x01, 0x00, 0x00,
}
