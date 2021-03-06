// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.13.0
// source: my_message.proto

package ld_proto

import (
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

type Feature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Geometry   *Feature_Geometry   `protobuf:"bytes,1,opt,name=geometry,proto3" json:"geometry,omitempty"`
	Properties *Feature_Properties `protobuf:"bytes,2,opt,name=properties,proto3" json:"properties,omitempty"`
	Type       string              `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
	Id         string              `protobuf:"bytes,4,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Feature) Reset() {
	*x = Feature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_my_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Feature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Feature) ProtoMessage() {}

func (x *Feature) ProtoReflect() protoreflect.Message {
	mi := &file_my_message_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Feature.ProtoReflect.Descriptor instead.
func (*Feature) Descriptor() ([]byte, []int) {
	return file_my_message_proto_rawDescGZIP(), []int{0}
}

func (x *Feature) GetGeometry() *Feature_Geometry {
	if x != nil {
		return x.Geometry
	}
	return nil
}

func (x *Feature) GetProperties() *Feature_Properties {
	if x != nil {
		return x.Properties
	}
	return nil
}

func (x *Feature) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Feature) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type Feature_Geometry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Coordinates []float64 `protobuf:"fixed64,1,rep,packed,name=coordinates,proto3" json:"coordinates,omitempty"`
	Type        string    `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *Feature_Geometry) Reset() {
	*x = Feature_Geometry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_my_message_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Feature_Geometry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Feature_Geometry) ProtoMessage() {}

func (x *Feature_Geometry) ProtoReflect() protoreflect.Message {
	mi := &file_my_message_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Feature_Geometry.ProtoReflect.Descriptor instead.
func (*Feature_Geometry) Descriptor() ([]byte, []int) {
	return file_my_message_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Feature_Geometry) GetCoordinates() []float64 {
	if x != nil {
		return x.Coordinates
	}
	return nil
}

func (x *Feature_Geometry) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

type Feature_Properties struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Amp      float32 `protobuf:"fixed32,1,opt,name=amp,proto3" json:"amp,omitempty"`
	Created  string  `protobuf:"bytes,2,opt,name=created,proto3" json:"created,omitempty"`
	Observed string  `protobuf:"bytes,3,opt,name=observed,proto3" json:"observed,omitempty"`
	Sensors  string  `protobuf:"bytes,4,opt,name=sensors,proto3" json:"sensors,omitempty"`
	Strokes  uint32  `protobuf:"varint,5,opt,name=strokes,proto3" json:"strokes,omitempty"`
	Type     uint32  `protobuf:"varint,6,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *Feature_Properties) Reset() {
	*x = Feature_Properties{}
	if protoimpl.UnsafeEnabled {
		mi := &file_my_message_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Feature_Properties) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Feature_Properties) ProtoMessage() {}

func (x *Feature_Properties) ProtoReflect() protoreflect.Message {
	mi := &file_my_message_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Feature_Properties.ProtoReflect.Descriptor instead.
func (*Feature_Properties) Descriptor() ([]byte, []int) {
	return file_my_message_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Feature_Properties) GetAmp() float32 {
	if x != nil {
		return x.Amp
	}
	return 0
}

func (x *Feature_Properties) GetCreated() string {
	if x != nil {
		return x.Created
	}
	return ""
}

func (x *Feature_Properties) GetObserved() string {
	if x != nil {
		return x.Observed
	}
	return ""
}

func (x *Feature_Properties) GetSensors() string {
	if x != nil {
		return x.Sensors
	}
	return ""
}

func (x *Feature_Properties) GetStrokes() uint32 {
	if x != nil {
		return x.Strokes
	}
	return 0
}

func (x *Feature_Properties) GetType() uint32 {
	if x != nil {
		return x.Type
	}
	return 0
}

var File_my_message_proto protoreflect.FileDescriptor

var file_my_message_proto_rawDesc = []byte{
	0x0a, 0x10, 0x6d, 0x79, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x08, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x84, 0x03, 0x0a,
	0x07, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x36, 0x0a, 0x08, 0x67, 0x65, 0x6f, 0x6d,
	0x65, 0x74, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6c, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2e, 0x47, 0x65,
	0x6f, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x52, 0x08, 0x67, 0x65, 0x6f, 0x6d, 0x65, 0x74, 0x72, 0x79,
	0x12, 0x3c, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69,
	0x65, 0x73, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x1a, 0x40, 0x0a, 0x08, 0x47, 0x65, 0x6f, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x12, 0x20,
	0x0a, 0x0b, 0x63, 0x6f, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x01, 0x52, 0x0b, 0x63, 0x6f, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x73,
	0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x1a, 0x9c, 0x01, 0x0a, 0x0a, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74,
	0x69, 0x65, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x02,
	0x52, 0x03, 0x61, 0x6d, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12,
	0x1a, 0x0a, 0x08, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x73,
	0x65, 0x6e, 0x73, 0x6f, 0x72, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x65,
	0x6e, 0x73, 0x6f, 0x72, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x72, 0x6f, 0x6b, 0x65, 0x73,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x73, 0x74, 0x72, 0x6f, 0x6b, 0x65, 0x73, 0x12,
	0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_my_message_proto_rawDescOnce sync.Once
	file_my_message_proto_rawDescData = file_my_message_proto_rawDesc
)

func file_my_message_proto_rawDescGZIP() []byte {
	file_my_message_proto_rawDescOnce.Do(func() {
		file_my_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_my_message_proto_rawDescData)
	})
	return file_my_message_proto_rawDescData
}

var file_my_message_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_my_message_proto_goTypes = []interface{}{
	(*Feature)(nil),            // 0: ld.proto.Feature
	(*Feature_Geometry)(nil),   // 1: ld.proto.Feature.Geometry
	(*Feature_Properties)(nil), // 2: ld.proto.Feature.Properties
}
var file_my_message_proto_depIdxs = []int32{
	1, // 0: ld.proto.Feature.geometry:type_name -> ld.proto.Feature.Geometry
	2, // 1: ld.proto.Feature.properties:type_name -> ld.proto.Feature.Properties
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_my_message_proto_init() }
func file_my_message_proto_init() {
	if File_my_message_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_my_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Feature); i {
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
		file_my_message_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Feature_Geometry); i {
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
		file_my_message_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Feature_Properties); i {
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
			RawDescriptor: file_my_message_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_my_message_proto_goTypes,
		DependencyIndexes: file_my_message_proto_depIdxs,
		MessageInfos:      file_my_message_proto_msgTypes,
	}.Build()
	File_my_message_proto = out.File
	file_my_message_proto_rawDesc = nil
	file_my_message_proto_goTypes = nil
	file_my_message_proto_depIdxs = nil
}
