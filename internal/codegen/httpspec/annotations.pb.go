// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: annotations.proto

package httpspec

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ProtoEndpointBinding struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Pattern:
	//
	//	*ProtoEndpointBinding_Get
	//	*ProtoEndpointBinding_Put
	//	*ProtoEndpointBinding_Post
	//	*ProtoEndpointBinding_Delete
	//	*ProtoEndpointBinding_Patch
	//	*ProtoEndpointBinding_Custom
	Pattern                    isProtoEndpointBinding_Pattern `protobuf_oneof:"pattern"`
	Body                       string                         `protobuf:"bytes,8,opt,name=body,proto3" json:"body,omitempty"`
	QueryParams                []*QueryParameterBinding       `protobuf:"bytes,9,rep,name=query_params,json=queryParams,proto3" json:"query_params,omitempty"`
	AdditionalBindings         []*AdditionalEndpointBinding   `protobuf:"bytes,10,rep,name=additional_bindings,json=additionalBindings,proto3" json:"additional_bindings,omitempty"`
	DisableQueryParamDiscovery bool                           `protobuf:"varint,11,opt,name=disable_query_param_discovery,json=disableQueryParamDiscovery,proto3" json:"disable_query_param_discovery,omitempty"`
}

func (x *ProtoEndpointBinding) Reset() {
	*x = ProtoEndpointBinding{}
	if protoimpl.UnsafeEnabled {
		mi := &file_annotations_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoEndpointBinding) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoEndpointBinding) ProtoMessage() {}

func (x *ProtoEndpointBinding) ProtoReflect() protoreflect.Message {
	mi := &file_annotations_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoEndpointBinding.ProtoReflect.Descriptor instead.
func (*ProtoEndpointBinding) Descriptor() ([]byte, []int) {
	return file_annotations_proto_rawDescGZIP(), []int{0}
}

func (m *ProtoEndpointBinding) GetPattern() isProtoEndpointBinding_Pattern {
	if m != nil {
		return m.Pattern
	}
	return nil
}

func (x *ProtoEndpointBinding) GetGet() string {
	if x, ok := x.GetPattern().(*ProtoEndpointBinding_Get); ok {
		return x.Get
	}
	return ""
}

func (x *ProtoEndpointBinding) GetPut() string {
	if x, ok := x.GetPattern().(*ProtoEndpointBinding_Put); ok {
		return x.Put
	}
	return ""
}

func (x *ProtoEndpointBinding) GetPost() string {
	if x, ok := x.GetPattern().(*ProtoEndpointBinding_Post); ok {
		return x.Post
	}
	return ""
}

func (x *ProtoEndpointBinding) GetDelete() string {
	if x, ok := x.GetPattern().(*ProtoEndpointBinding_Delete); ok {
		return x.Delete
	}
	return ""
}

func (x *ProtoEndpointBinding) GetPatch() string {
	if x, ok := x.GetPattern().(*ProtoEndpointBinding_Patch); ok {
		return x.Patch
	}
	return ""
}

func (x *ProtoEndpointBinding) GetCustom() *CustomPattern {
	if x, ok := x.GetPattern().(*ProtoEndpointBinding_Custom); ok {
		return x.Custom
	}
	return nil
}

func (x *ProtoEndpointBinding) GetBody() string {
	if x != nil {
		return x.Body
	}
	return ""
}

func (x *ProtoEndpointBinding) GetQueryParams() []*QueryParameterBinding {
	if x != nil {
		return x.QueryParams
	}
	return nil
}

func (x *ProtoEndpointBinding) GetAdditionalBindings() []*AdditionalEndpointBinding {
	if x != nil {
		return x.AdditionalBindings
	}
	return nil
}

func (x *ProtoEndpointBinding) GetDisableQueryParamDiscovery() bool {
	if x != nil {
		return x.DisableQueryParamDiscovery
	}
	return false
}

type isProtoEndpointBinding_Pattern interface {
	isProtoEndpointBinding_Pattern()
}

type ProtoEndpointBinding_Get struct {
	Get string `protobuf:"bytes,2,opt,name=get,proto3,oneof"`
}

type ProtoEndpointBinding_Put struct {
	Put string `protobuf:"bytes,3,opt,name=put,proto3,oneof"`
}

type ProtoEndpointBinding_Post struct {
	Post string `protobuf:"bytes,4,opt,name=post,proto3,oneof"`
}

type ProtoEndpointBinding_Delete struct {
	Delete string `protobuf:"bytes,5,opt,name=delete,proto3,oneof"`
}

type ProtoEndpointBinding_Patch struct {
	Patch string `protobuf:"bytes,6,opt,name=patch,proto3,oneof"`
}

type ProtoEndpointBinding_Custom struct {
	// custom can be used for custom HTTP methods.
	Custom *CustomPattern `protobuf:"bytes,7,opt,name=custom,proto3,oneof"`
}

func (*ProtoEndpointBinding_Get) isProtoEndpointBinding_Pattern() {}

func (*ProtoEndpointBinding_Put) isProtoEndpointBinding_Pattern() {}

func (*ProtoEndpointBinding_Post) isProtoEndpointBinding_Pattern() {}

func (*ProtoEndpointBinding_Delete) isProtoEndpointBinding_Pattern() {}

func (*ProtoEndpointBinding_Patch) isProtoEndpointBinding_Pattern() {}

func (*ProtoEndpointBinding_Custom) isProtoEndpointBinding_Pattern() {}

var file_annotations_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*ProtoEndpointBinding)(nil),
		Field:         1142,
		Name:          "meshapi.gateway.http",
		Tag:           "bytes,1142,opt,name=http",
		Filename:      "annotations.proto",
	},
}

// Extension fields to descriptorpb.MethodOptions.
var (
	// http holds HTTP endpoint binding configs.
	//
	// ID assigned by protobuf-global-extension-registry@google.com for gRPC-Gateway project.
	//
	// optional meshapi.gateway.ProtoEndpointBinding http = 1142;
	E_Http = &file_annotations_proto_extTypes[0]
)

var File_annotations_proto protoreflect.FileDescriptor

var file_annotations_proto_rawDesc = []byte{
	0x0a, 0x11, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2e, 0x67, 0x61, 0x74,
	0x65, 0x77, 0x61, 0x79, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xca, 0x03, 0x0a, 0x14, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45,
	0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x42, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x12,
	0x0a, 0x03, 0x67, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x03, 0x67,
	0x65, 0x74, 0x12, 0x12, 0x0a, 0x03, 0x70, 0x75, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48,
	0x00, 0x52, 0x03, 0x70, 0x75, 0x74, 0x12, 0x14, 0x0a, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x06,
	0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06,
	0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x05, 0x70, 0x61, 0x74, 0x63, 0x68, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x70, 0x61, 0x74, 0x63, 0x68, 0x12, 0x38,
	0x0a, 0x06, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2e, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x2e, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x50, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x48, 0x00,
	0x52, 0x06, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x49, 0x0a, 0x0c,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x18, 0x09, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x26, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2e, 0x67, 0x61, 0x74,
	0x65, 0x77, 0x61, 0x79, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65,
	0x74, 0x65, 0x72, 0x42, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x52, 0x0b, 0x71, 0x75, 0x65, 0x72,
	0x79, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x5b, 0x0a, 0x13, 0x61, 0x64, 0x64, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x62, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x0a,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2e, 0x67,
	0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x41, 0x64, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x61,
	0x6c, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x42, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67,
	0x52, 0x12, 0x61, 0x64, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x42, 0x69, 0x6e, 0x64,
	0x69, 0x6e, 0x67, 0x73, 0x12, 0x41, 0x0a, 0x1d, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x5f,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x5f, 0x64, 0x69, 0x73, 0x63,
	0x6f, 0x76, 0x65, 0x72, 0x79, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x08, 0x52, 0x1a, 0x64, 0x69, 0x73,
	0x61, 0x62, 0x6c, 0x65, 0x51, 0x75, 0x65, 0x72, 0x79, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x44, 0x69,
	0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65,
	0x72, 0x6e, 0x3a, 0x5a, 0x0a, 0x04, 0x68, 0x74, 0x74, 0x70, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xf6, 0x08, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x25, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2e, 0x67, 0x61, 0x74, 0x65,
	0x77, 0x61, 0x79, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x42, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x52, 0x04, 0x68, 0x74, 0x74, 0x70, 0x42, 0x40,
	0x5a, 0x3e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x65, 0x73,
	0x68, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2d, 0x72, 0x65, 0x73, 0x74, 0x2d, 0x67,
	0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f,
	0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x2f, 0x68, 0x74, 0x74, 0x70, 0x73, 0x70, 0x65, 0x63,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_annotations_proto_rawDescOnce sync.Once
	file_annotations_proto_rawDescData = file_annotations_proto_rawDesc
)

func file_annotations_proto_rawDescGZIP() []byte {
	file_annotations_proto_rawDescOnce.Do(func() {
		file_annotations_proto_rawDescData = protoimpl.X.CompressGZIP(file_annotations_proto_rawDescData)
	})
	return file_annotations_proto_rawDescData
}

var file_annotations_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_annotations_proto_goTypes = []interface{}{
	(*ProtoEndpointBinding)(nil),       // 0: meshapi.gateway.ProtoEndpointBinding
	(*CustomPattern)(nil),              // 1: meshapi.gateway.CustomPattern
	(*QueryParameterBinding)(nil),      // 2: meshapi.gateway.QueryParameterBinding
	(*AdditionalEndpointBinding)(nil),  // 3: meshapi.gateway.AdditionalEndpointBinding
	(*descriptorpb.MethodOptions)(nil), // 4: google.protobuf.MethodOptions
}
var file_annotations_proto_depIdxs = []int32{
	1, // 0: meshapi.gateway.ProtoEndpointBinding.custom:type_name -> meshapi.gateway.CustomPattern
	2, // 1: meshapi.gateway.ProtoEndpointBinding.query_params:type_name -> meshapi.gateway.QueryParameterBinding
	3, // 2: meshapi.gateway.ProtoEndpointBinding.additional_bindings:type_name -> meshapi.gateway.AdditionalEndpointBinding
	4, // 3: meshapi.gateway.http:extendee -> google.protobuf.MethodOptions
	0, // 4: meshapi.gateway.http:type_name -> meshapi.gateway.ProtoEndpointBinding
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	4, // [4:5] is the sub-list for extension type_name
	3, // [3:4] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_annotations_proto_init() }
func file_annotations_proto_init() {
	if File_annotations_proto != nil {
		return
	}
	file_gateway_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_annotations_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoEndpointBinding); i {
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
	file_annotations_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*ProtoEndpointBinding_Get)(nil),
		(*ProtoEndpointBinding_Put)(nil),
		(*ProtoEndpointBinding_Post)(nil),
		(*ProtoEndpointBinding_Delete)(nil),
		(*ProtoEndpointBinding_Patch)(nil),
		(*ProtoEndpointBinding_Custom)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_annotations_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 1,
			NumServices:   0,
		},
		GoTypes:           file_annotations_proto_goTypes,
		DependencyIndexes: file_annotations_proto_depIdxs,
		MessageInfos:      file_annotations_proto_msgTypes,
		ExtensionInfos:    file_annotations_proto_extTypes,
	}.Build()
	File_annotations_proto = out.File
	file_annotations_proto_rawDesc = nil
	file_annotations_proto_goTypes = nil
	file_annotations_proto_depIdxs = nil
}
