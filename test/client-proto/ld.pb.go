// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.13.0
// source: ld.proto

package ld_proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

//The Key when querying directly for it
//The Key in general could be any bytes, but pattern-scanning requires string,
//so I have decided to increase the requirements in order to add the convenience
//of pattern-searching.
type Key struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *Key) Reset() {
	*x = Key{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ld_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Key) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Key) ProtoMessage() {}

func (x *Key) ProtoReflect() protoreflect.Message {
	mi := &file_ld_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Key.ProtoReflect.Descriptor instead.
func (*Key) Descriptor() ([]byte, []int) {
	return file_ld_proto_rawDescGZIP(), []int{0}
}

func (x *Key) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

//A key-range is the only possibility of querying the data outside of a direct Key.
//The logical operator between using prefix, pattern and from-to together is AND.
//OR is not implemented as it can be done using more than one request
//Empty KeyRange implies a full database stream
type KeyRange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//A key-prefix to search within.
	//when using prefix along-side pattern and/or from-to they should both match.
	// ie. a prefix "jo" could be used to speed up query speed of
	//     pattern "john*" or from: "john1" to: "john6"
	//the server will not try to guess a prefix from the pattern or from-to parameters
	//pattern-searching is the slowest operation.
	//pattern john* is the same as prefix: john, but slower
	Prefix string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	// RE2 style regex syntax via golang core: https://golang.org/pkg/regexp/
	Pattern string `protobuf:"bytes,2,opt,name=pattern,proto3" json:"pattern,omitempty"`
	// both inclusive
	// required for discrete systems with discrete queries
	//  -- since you cannot reference a value outside of the last/first,
	//     and would then not be able to query the last/first record.
	//     and +1 semantics on strings don't really work
	From string `protobuf:"bytes,3,opt,name=from,proto3" json:"from,omitempty"`
	To   string `protobuf:"bytes,4,opt,name=to,proto3" json:"to,omitempty"`
}

func (x *KeyRange) Reset() {
	*x = KeyRange{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ld_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyRange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyRange) ProtoMessage() {}

func (x *KeyRange) ProtoReflect() protoreflect.Message {
	mi := &file_ld_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeyRange.ProtoReflect.Descriptor instead.
func (*KeyRange) Descriptor() ([]byte, []int) {
	return file_ld_proto_rawDescGZIP(), []int{1}
}

func (x *KeyRange) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *KeyRange) GetPattern() string {
	if x != nil {
		return x.Pattern
	}
	return ""
}

func (x *KeyRange) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *KeyRange) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

type KeyValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	//You can easily replace this with google's Any if your want to
	//Or replace with your own message-type
	//
	//fx you have some software that simply expose data from a datasource
	//Your software exposes it as proto. This will be your datasource.
	// rewrite this .proto-file on the client side
	// add `import "your_messages_file.proto"`
	// replace the bytes of this with the type/format you wish to save
	// this works because string, bytes and nested messages are encoded the same:
	//   read https://developers.google.com/protocol-buffers/docs/encoding#strings
	//   and https://developers.google.com/protocol-buffers/docs/encoding#embedded
	Value *Feature `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *KeyValue) Reset() {
	*x = KeyValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ld_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyValue) ProtoMessage() {}

func (x *KeyValue) ProtoReflect() protoreflect.Message {
	mi := &file_ld_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeyValue.ProtoReflect.Descriptor instead.
func (*KeyValue) Descriptor() ([]byte, []int) {
	return file_ld_proto_rawDescGZIP(), []int{2}
}

func (x *KeyValue) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *KeyValue) GetValue() *Feature {
	if x != nil {
		return x.Value
	}
	return nil
}

var File_ld_proto protoreflect.FileDescriptor

var file_ld_proto_rawDesc = []byte{
	0x0a, 0x08, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x6c, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x6d, 0x79, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x17, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22,
	0x60, 0x0a, 0x08, 0x4b, 0x65, 0x79, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x70,
	0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72, 0x65,
	0x66, 0x69, 0x78, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x72, 0x6f,
	0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74,
	0x6f, 0x22, 0x45, 0x0a, 0x08, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x27, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11,
	0x2e, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0x97, 0x03, 0x0a, 0x02, 0x6c, 0x64, 0x12,
	0x2d, 0x0a, 0x03, 0x53, 0x65, 0x74, 0x12, 0x12, 0x2e, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x12, 0x2e, 0x6c, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x35,
	0x0a, 0x07, 0x53, 0x65, 0x74, 0x4d, 0x61, 0x6e, 0x79, 0x12, 0x12, 0x2e, 0x6c, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x12, 0x2e,
	0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x28, 0x01, 0x30, 0x01, 0x12, 0x28, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x0d, 0x2e, 0x6c,
	0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x1a, 0x12, 0x2e, 0x6c, 0x64,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x30, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x4d, 0x61, 0x6e, 0x79, 0x12, 0x0d, 0x2e, 0x6c, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x1a, 0x12, 0x2e, 0x6c, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x28, 0x01, 0x30,
	0x01, 0x12, 0x34, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x12, 0x2e,
	0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x52, 0x61, 0x6e, 0x67,
	0x65, 0x1a, 0x12, 0x2e, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x30, 0x01, 0x12, 0x2b, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x12, 0x0d, 0x2e, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79,
	0x1a, 0x12, 0x2e, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x12, 0x33, 0x0a, 0x0a, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4d, 0x61,
	0x6e, 0x79, 0x12, 0x0d, 0x2e, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65,
	0x79, 0x1a, 0x12, 0x2e, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x28, 0x01, 0x30, 0x01, 0x12, 0x37, 0x0a, 0x0b, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x12, 0x2e, 0x6c, 0x64, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x1a, 0x12, 0x2e, 0x6c,
	0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x30, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ld_proto_rawDescOnce sync.Once
	file_ld_proto_rawDescData = file_ld_proto_rawDesc
)

func file_ld_proto_rawDescGZIP() []byte {
	file_ld_proto_rawDescOnce.Do(func() {
		file_ld_proto_rawDescData = protoimpl.X.CompressGZIP(file_ld_proto_rawDescData)
	})
	return file_ld_proto_rawDescData
}

var file_ld_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_ld_proto_goTypes = []interface{}{
	(*Key)(nil),      // 0: ld.proto.Key
	(*KeyRange)(nil), // 1: ld.proto.KeyRange
	(*KeyValue)(nil), // 2: ld.proto.KeyValue
	(*Feature)(nil),  // 3: ld.proto.Feature
}
var file_ld_proto_depIdxs = []int32{
	3, // 0: ld.proto.KeyValue.value:type_name -> ld.proto.Feature
	2, // 1: ld.proto.ld.Set:input_type -> ld.proto.KeyValue
	2, // 2: ld.proto.ld.SetMany:input_type -> ld.proto.KeyValue
	0, // 3: ld.proto.ld.Get:input_type -> ld.proto.Key
	0, // 4: ld.proto.ld.GetMany:input_type -> ld.proto.Key
	1, // 5: ld.proto.ld.GetRange:input_type -> ld.proto.KeyRange
	0, // 6: ld.proto.ld.Delete:input_type -> ld.proto.Key
	0, // 7: ld.proto.ld.DeleteMany:input_type -> ld.proto.Key
	1, // 8: ld.proto.ld.DeleteRange:input_type -> ld.proto.KeyRange
	2, // 9: ld.proto.ld.Set:output_type -> ld.proto.KeyValue
	2, // 10: ld.proto.ld.SetMany:output_type -> ld.proto.KeyValue
	2, // 11: ld.proto.ld.Get:output_type -> ld.proto.KeyValue
	2, // 12: ld.proto.ld.GetMany:output_type -> ld.proto.KeyValue
	2, // 13: ld.proto.ld.GetRange:output_type -> ld.proto.KeyValue
	2, // 14: ld.proto.ld.Delete:output_type -> ld.proto.KeyValue
	2, // 15: ld.proto.ld.DeleteMany:output_type -> ld.proto.KeyValue
	2, // 16: ld.proto.ld.DeleteRange:output_type -> ld.proto.KeyValue
	9, // [9:17] is the sub-list for method output_type
	1, // [1:9] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_ld_proto_init() }
func file_ld_proto_init() {
	if File_ld_proto != nil {
		return
	}
	file_my_message_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_ld_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Key); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ld_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeyRange); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ld_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeyValue); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ld_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ld_proto_goTypes,
		DependencyIndexes: file_ld_proto_depIdxs,
		MessageInfos:      file_ld_proto_msgTypes,
	}.Build()
	File_ld_proto = out.File
	file_ld_proto_rawDesc = nil
	file_ld_proto_goTypes = nil
	file_ld_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// LdClient is the client API for Ld service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type LdClient interface {
	//empty response means success
	//the database returns your KeyValue for errors
	Set(ctx context.Context, in *KeyValue, opts ...grpc.CallOption) (*KeyValue, error)
	SetMany(ctx context.Context, opts ...grpc.CallOption) (Ld_SetManyClient, error)
	//empty responses means no such key.
	Get(ctx context.Context, in *Key, opts ...grpc.CallOption) (*KeyValue, error)
	GetMany(ctx context.Context, opts ...grpc.CallOption) (Ld_GetManyClient, error)
	GetRange(ctx context.Context, in *KeyRange, opts ...grpc.CallOption) (Ld_GetRangeClient, error)
	//returns the deleted object, empty means no such key
	Delete(ctx context.Context, in *Key, opts ...grpc.CallOption) (*KeyValue, error)
	DeleteMany(ctx context.Context, opts ...grpc.CallOption) (Ld_DeleteManyClient, error)
	DeleteRange(ctx context.Context, in *KeyRange, opts ...grpc.CallOption) (Ld_DeleteRangeClient, error)
}

type ldClient struct {
	cc grpc.ClientConnInterface
}

func NewLdClient(cc grpc.ClientConnInterface) LdClient {
	return &ldClient{cc}
}

func (c *ldClient) Set(ctx context.Context, in *KeyValue, opts ...grpc.CallOption) (*KeyValue, error) {
	out := new(KeyValue)
	err := c.cc.Invoke(ctx, "/ld.proto.ld/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ldClient) SetMany(ctx context.Context, opts ...grpc.CallOption) (Ld_SetManyClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Ld_serviceDesc.Streams[0], "/ld.proto.ld/SetMany", opts...)
	if err != nil {
		return nil, err
	}
	x := &ldSetManyClient{stream}
	return x, nil
}

type Ld_SetManyClient interface {
	Send(*KeyValue) error
	Recv() (*KeyValue, error)
	grpc.ClientStream
}

type ldSetManyClient struct {
	grpc.ClientStream
}

func (x *ldSetManyClient) Send(m *KeyValue) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ldSetManyClient) Recv() (*KeyValue, error) {
	m := new(KeyValue)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ldClient) Get(ctx context.Context, in *Key, opts ...grpc.CallOption) (*KeyValue, error) {
	out := new(KeyValue)
	err := c.cc.Invoke(ctx, "/ld.proto.ld/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ldClient) GetMany(ctx context.Context, opts ...grpc.CallOption) (Ld_GetManyClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Ld_serviceDesc.Streams[1], "/ld.proto.ld/GetMany", opts...)
	if err != nil {
		return nil, err
	}
	x := &ldGetManyClient{stream}
	return x, nil
}

type Ld_GetManyClient interface {
	Send(*Key) error
	Recv() (*KeyValue, error)
	grpc.ClientStream
}

type ldGetManyClient struct {
	grpc.ClientStream
}

func (x *ldGetManyClient) Send(m *Key) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ldGetManyClient) Recv() (*KeyValue, error) {
	m := new(KeyValue)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ldClient) GetRange(ctx context.Context, in *KeyRange, opts ...grpc.CallOption) (Ld_GetRangeClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Ld_serviceDesc.Streams[2], "/ld.proto.ld/GetRange", opts...)
	if err != nil {
		return nil, err
	}
	x := &ldGetRangeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Ld_GetRangeClient interface {
	Recv() (*KeyValue, error)
	grpc.ClientStream
}

type ldGetRangeClient struct {
	grpc.ClientStream
}

func (x *ldGetRangeClient) Recv() (*KeyValue, error) {
	m := new(KeyValue)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ldClient) Delete(ctx context.Context, in *Key, opts ...grpc.CallOption) (*KeyValue, error) {
	out := new(KeyValue)
	err := c.cc.Invoke(ctx, "/ld.proto.ld/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ldClient) DeleteMany(ctx context.Context, opts ...grpc.CallOption) (Ld_DeleteManyClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Ld_serviceDesc.Streams[3], "/ld.proto.ld/DeleteMany", opts...)
	if err != nil {
		return nil, err
	}
	x := &ldDeleteManyClient{stream}
	return x, nil
}

type Ld_DeleteManyClient interface {
	Send(*Key) error
	Recv() (*KeyValue, error)
	grpc.ClientStream
}

type ldDeleteManyClient struct {
	grpc.ClientStream
}

func (x *ldDeleteManyClient) Send(m *Key) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ldDeleteManyClient) Recv() (*KeyValue, error) {
	m := new(KeyValue)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ldClient) DeleteRange(ctx context.Context, in *KeyRange, opts ...grpc.CallOption) (Ld_DeleteRangeClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Ld_serviceDesc.Streams[4], "/ld.proto.ld/DeleteRange", opts...)
	if err != nil {
		return nil, err
	}
	x := &ldDeleteRangeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Ld_DeleteRangeClient interface {
	Recv() (*KeyValue, error)
	grpc.ClientStream
}

type ldDeleteRangeClient struct {
	grpc.ClientStream
}

func (x *ldDeleteRangeClient) Recv() (*KeyValue, error) {
	m := new(KeyValue)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// LdServer is the server API for Ld service.
type LdServer interface {
	//empty response means success
	//the database returns your KeyValue for errors
	Set(context.Context, *KeyValue) (*KeyValue, error)
	SetMany(Ld_SetManyServer) error
	//empty responses means no such key.
	Get(context.Context, *Key) (*KeyValue, error)
	GetMany(Ld_GetManyServer) error
	GetRange(*KeyRange, Ld_GetRangeServer) error
	//returns the deleted object, empty means no such key
	Delete(context.Context, *Key) (*KeyValue, error)
	DeleteMany(Ld_DeleteManyServer) error
	DeleteRange(*KeyRange, Ld_DeleteRangeServer) error
}

// UnimplementedLdServer can be embedded to have forward compatible implementations.
type UnimplementedLdServer struct {
}

func (*UnimplementedLdServer) Set(context.Context, *KeyValue) (*KeyValue, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (*UnimplementedLdServer) SetMany(Ld_SetManyServer) error {
	return status.Errorf(codes.Unimplemented, "method SetMany not implemented")
}
func (*UnimplementedLdServer) Get(context.Context, *Key) (*KeyValue, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (*UnimplementedLdServer) GetMany(Ld_GetManyServer) error {
	return status.Errorf(codes.Unimplemented, "method GetMany not implemented")
}
func (*UnimplementedLdServer) GetRange(*KeyRange, Ld_GetRangeServer) error {
	return status.Errorf(codes.Unimplemented, "method GetRange not implemented")
}
func (*UnimplementedLdServer) Delete(context.Context, *Key) (*KeyValue, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (*UnimplementedLdServer) DeleteMany(Ld_DeleteManyServer) error {
	return status.Errorf(codes.Unimplemented, "method DeleteMany not implemented")
}
func (*UnimplementedLdServer) DeleteRange(*KeyRange, Ld_DeleteRangeServer) error {
	return status.Errorf(codes.Unimplemented, "method DeleteRange not implemented")
}

func RegisterLdServer(s *grpc.Server, srv LdServer) {
	s.RegisterService(&_Ld_serviceDesc, srv)
}

func _Ld_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KeyValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LdServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ld.proto.ld/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LdServer).Set(ctx, req.(*KeyValue))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ld_SetMany_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LdServer).SetMany(&ldSetManyServer{stream})
}

type Ld_SetManyServer interface {
	Send(*KeyValue) error
	Recv() (*KeyValue, error)
	grpc.ServerStream
}

type ldSetManyServer struct {
	grpc.ServerStream
}

func (x *ldSetManyServer) Send(m *KeyValue) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ldSetManyServer) Recv() (*KeyValue, error) {
	m := new(KeyValue)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Ld_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Key)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LdServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ld.proto.ld/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LdServer).Get(ctx, req.(*Key))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ld_GetMany_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LdServer).GetMany(&ldGetManyServer{stream})
}

type Ld_GetManyServer interface {
	Send(*KeyValue) error
	Recv() (*Key, error)
	grpc.ServerStream
}

type ldGetManyServer struct {
	grpc.ServerStream
}

func (x *ldGetManyServer) Send(m *KeyValue) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ldGetManyServer) Recv() (*Key, error) {
	m := new(Key)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Ld_GetRange_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(KeyRange)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(LdServer).GetRange(m, &ldGetRangeServer{stream})
}

type Ld_GetRangeServer interface {
	Send(*KeyValue) error
	grpc.ServerStream
}

type ldGetRangeServer struct {
	grpc.ServerStream
}

func (x *ldGetRangeServer) Send(m *KeyValue) error {
	return x.ServerStream.SendMsg(m)
}

func _Ld_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Key)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LdServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ld.proto.ld/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LdServer).Delete(ctx, req.(*Key))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ld_DeleteMany_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LdServer).DeleteMany(&ldDeleteManyServer{stream})
}

type Ld_DeleteManyServer interface {
	Send(*KeyValue) error
	Recv() (*Key, error)
	grpc.ServerStream
}

type ldDeleteManyServer struct {
	grpc.ServerStream
}

func (x *ldDeleteManyServer) Send(m *KeyValue) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ldDeleteManyServer) Recv() (*Key, error) {
	m := new(Key)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Ld_DeleteRange_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(KeyRange)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(LdServer).DeleteRange(m, &ldDeleteRangeServer{stream})
}

type Ld_DeleteRangeServer interface {
	Send(*KeyValue) error
	grpc.ServerStream
}

type ldDeleteRangeServer struct {
	grpc.ServerStream
}

func (x *ldDeleteRangeServer) Send(m *KeyValue) error {
	return x.ServerStream.SendMsg(m)
}

var _Ld_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ld.proto.ld",
	HandlerType: (*LdServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Set",
			Handler:    _Ld_Set_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _Ld_Get_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _Ld_Delete_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SetMany",
			Handler:       _Ld_SetMany_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "GetMany",
			Handler:       _Ld_GetMany_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "GetRange",
			Handler:       _Ld_GetRange_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "DeleteMany",
			Handler:       _Ld_DeleteMany_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "DeleteRange",
			Handler:       _Ld_DeleteRange_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "ld.proto",
}
