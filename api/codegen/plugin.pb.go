// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: meshapi/gateway/codegen/plugin.proto

package codegen

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

type Generator int32

const (
	Generator_Generator_UNKNOWN     Generator = 0
	Generator_Generator_RestGateway Generator = 1
	Generator_Generator_OpenAPI     Generator = 2
)

// Enum value maps for Generator.
var (
	Generator_name = map[int32]string{
		0: "Generator_UNKNOWN",
		1: "Generator_RestGateway",
		2: "Generator_OpenAPI",
	}
	Generator_value = map[string]int32{
		"Generator_UNKNOWN":     0,
		"Generator_RestGateway": 1,
		"Generator_OpenAPI":     2,
	}
)

func (x Generator) Enum() *Generator {
	p := new(Generator)
	*p = x
	return p
}

func (x Generator) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Generator) Descriptor() protoreflect.EnumDescriptor {
	return file_meshapi_gateway_codegen_plugin_proto_enumTypes[0].Descriptor()
}

func (Generator) Type() protoreflect.EnumType {
	return &file_meshapi_gateway_codegen_plugin_proto_enumTypes[0]
}

func (x Generator) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Generator.Descriptor instead.
func (Generator) EnumDescriptor() ([]byte, []int) {
	return file_meshapi_gateway_codegen_plugin_proto_rawDescGZIP(), []int{0}
}

// UnixSocketConnection is a mode of plugin connectivity using UNIX sockets.
type UnixSocketConnection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// socket is the socket file.
	Socket string `protobuf:"bytes,1,opt,name=socket,proto3" json:"socket,omitempty"`
}

func (x *UnixSocketConnection) Reset() {
	*x = UnixSocketConnection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UnixSocketConnection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnixSocketConnection) ProtoMessage() {}

func (x *UnixSocketConnection) ProtoReflect() protoreflect.Message {
	mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnixSocketConnection.ProtoReflect.Descriptor instead.
func (*UnixSocketConnection) Descriptor() ([]byte, []int) {
	return file_meshapi_gateway_codegen_plugin_proto_rawDescGZIP(), []int{0}
}

func (x *UnixSocketConnection) GetSocket() string {
	if x != nil {
		return x.Socket
	}
	return ""
}

// TCPConnection is a mode of plugin connectivity via TCP.
type TCPConnection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// address is the TCP connection address to use to connect to the plugin.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *TCPConnection) Reset() {
	*x = TCPConnection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TCPConnection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TCPConnection) ProtoMessage() {}

func (x *TCPConnection) ProtoReflect() protoreflect.Message {
	mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TCPConnection.ProtoReflect.Descriptor instead.
func (*TCPConnection) Descriptor() ([]byte, []int) {
	return file_meshapi_gateway_codegen_plugin_proto_rawDescGZIP(), []int{1}
}

func (x *TCPConnection) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

// Version describes a semver version.
type Version struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// major is the version's major segment.
	Major uint32 `protobuf:"varint,1,opt,name=major,proto3" json:"major,omitempty"`
	// minor is the version's minor segment.
	Minor uint32 `protobuf:"varint,2,opt,name=minor,proto3" json:"minor,omitempty"`
	// patch is the version's patch segment.
	Patch uint32 `protobuf:"varint,3,opt,name=patch,proto3" json:"patch,omitempty"`
}

func (x *Version) Reset() {
	*x = Version{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Version) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Version) ProtoMessage() {}

func (x *Version) ProtoReflect() protoreflect.Message {
	mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Version.ProtoReflect.Descriptor instead.
func (*Version) Descriptor() ([]byte, []int) {
	return file_meshapi_gateway_codegen_plugin_proto_rawDescGZIP(), []int{2}
}

func (x *Version) GetMajor() uint32 {
	if x != nil {
		return x.Major
	}
	return 0
}

func (x *Version) GetMinor() uint32 {
	if x != nil {
		return x.Minor
	}
	return 0
}

func (x *Version) GetPatch() uint32 {
	if x != nil {
		return x.Patch
	}
	return 0
}

// GeneratorInfo is the first message that is written in the plugin process's stdin
// which dumps some details about the generator's versions and supported callbacks.
type GeneratorInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// version indicates the code generator's version.
	Version *Version `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	// generator is the type of generator activating this plugin.
	Generator Generator `protobuf:"varint,2,opt,name=generator,proto3,enum=meshapi.gateway.codegen.Generator" json:"generator,omitempty"`
	// supported_features are all available features in the code generator.
	SupportedFeatures []string `protobuf:"bytes,3,rep,name=supported_features,json=supportedFeatures,proto3" json:"supported_features,omitempty"`
}

func (x *GeneratorInfo) Reset() {
	*x = GeneratorInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeneratorInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeneratorInfo) ProtoMessage() {}

func (x *GeneratorInfo) ProtoReflect() protoreflect.Message {
	mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeneratorInfo.ProtoReflect.Descriptor instead.
func (*GeneratorInfo) Descriptor() ([]byte, []int) {
	return file_meshapi_gateway_codegen_plugin_proto_rawDescGZIP(), []int{3}
}

func (x *GeneratorInfo) GetVersion() *Version {
	if x != nil {
		return x.Version
	}
	return nil
}

func (x *GeneratorInfo) GetGenerator() Generator {
	if x != nil {
		return x.Generator
	}
	return Generator_Generator_UNKNOWN
}

func (x *GeneratorInfo) GetSupportedFeatures() []string {
	if x != nil {
		return x.SupportedFeatures
	}
	return nil
}

// PluginInfo is the first message that is read from the process's output and must contain information about the
// plugin.
type PluginInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// address is the gRPC server address to use for connecting to the plugin.
	//
	// Types that are assignable to Connection:
	//
	//	*PluginInfo_UnixSocket
	//	*PluginInfo_Tcp
	Connection isPluginInfo_Connection `protobuf_oneof:"connection"`
	// registered_callbacks are all the callbacks the plug-in wishes to handle.
	RegisteredCallbacks []string `protobuf:"bytes,3,rep,name=registered_callbacks,json=registeredCallbacks,proto3" json:"registered_callbacks,omitempty"`
}

func (x *PluginInfo) Reset() {
	*x = PluginInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PluginInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PluginInfo) ProtoMessage() {}

func (x *PluginInfo) ProtoReflect() protoreflect.Message {
	mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PluginInfo.ProtoReflect.Descriptor instead.
func (*PluginInfo) Descriptor() ([]byte, []int) {
	return file_meshapi_gateway_codegen_plugin_proto_rawDescGZIP(), []int{4}
}

func (m *PluginInfo) GetConnection() isPluginInfo_Connection {
	if m != nil {
		return m.Connection
	}
	return nil
}

func (x *PluginInfo) GetUnixSocket() *UnixSocketConnection {
	if x, ok := x.GetConnection().(*PluginInfo_UnixSocket); ok {
		return x.UnixSocket
	}
	return nil
}

func (x *PluginInfo) GetTcp() *TCPConnection {
	if x, ok := x.GetConnection().(*PluginInfo_Tcp); ok {
		return x.Tcp
	}
	return nil
}

func (x *PluginInfo) GetRegisteredCallbacks() []string {
	if x != nil {
		return x.RegisteredCallbacks
	}
	return nil
}

type isPluginInfo_Connection interface {
	isPluginInfo_Connection()
}

type PluginInfo_UnixSocket struct {
	// socket is the connection via UNIX sockets.
	UnixSocket *UnixSocketConnection `protobuf:"bytes,1,opt,name=unix_socket,json=unixSocket,proto3,oneof"`
}

type PluginInfo_Tcp struct {
	// tcp is connection via TCP, useful in operating systems such as Windows where UNIX sockets are not available.
	Tcp *TCPConnection `protobuf:"bytes,2,opt,name=tcp,proto3,oneof"`
}

func (*PluginInfo_UnixSocket) isPluginInfo_Connection() {}

func (*PluginInfo_Tcp) isPluginInfo_Connection() {}

type PingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Text string `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
}

func (x *PingRequest) Reset() {
	*x = PingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingRequest) ProtoMessage() {}

func (x *PingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingRequest.ProtoReflect.Descriptor instead.
func (*PingRequest) Descriptor() ([]byte, []int) {
	return file_meshapi_gateway_codegen_plugin_proto_rawDescGZIP(), []int{5}
}

func (x *PingRequest) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

type PingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Text string `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingResponse) ProtoMessage() {}

func (x *PingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_meshapi_gateway_codegen_plugin_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingResponse.ProtoReflect.Descriptor instead.
func (*PingResponse) Descriptor() ([]byte, []int) {
	return file_meshapi_gateway_codegen_plugin_proto_rawDescGZIP(), []int{6}
}

func (x *PingResponse) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

var File_meshapi_gateway_codegen_plugin_proto protoreflect.FileDescriptor

var file_meshapi_gateway_codegen_plugin_proto_rawDesc = []byte{
	0x0a, 0x24, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61,
	0x79, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2e,
	0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x22,
	0x2e, 0x0a, 0x14, 0x55, 0x6e, 0x69, 0x78, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x43, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x63, 0x6b, 0x65,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x22,
	0x29, 0x0a, 0x0d, 0x54, 0x43, 0x50, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x4b, 0x0a, 0x07, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x61, 0x6a, 0x6f, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x6d, 0x61, 0x6a, 0x6f, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x6d,
	0x69, 0x6e, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x6d, 0x69, 0x6e, 0x6f,
	0x72, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x61, 0x74, 0x63, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x05, 0x70, 0x61, 0x74, 0x63, 0x68, 0x22, 0xbc, 0x01, 0x0a, 0x0d, 0x47, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x3a, 0x0a, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x61, 0x70, 0x69, 0x2e, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x64,
	0x65, 0x67, 0x65, 0x6e, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x40, 0x0a, 0x09, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x22, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61,
	0x70, 0x69, 0x2e, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x67,
	0x65, 0x6e, 0x2e, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x09, 0x67, 0x65,
	0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x2d, 0x0a, 0x12, 0x73, 0x75, 0x70, 0x70, 0x6f,
	0x72, 0x74, 0x65, 0x64, 0x5f, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x11, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x46, 0x65,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x22, 0xdb, 0x01, 0x0a, 0x0a, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x50, 0x0a, 0x0b, 0x75, 0x6e, 0x69, 0x78, 0x5f, 0x73, 0x6f,
	0x63, 0x6b, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x6d, 0x65, 0x73,
	0x68, 0x61, 0x70, 0x69, 0x2e, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x64,
	0x65, 0x67, 0x65, 0x6e, 0x2e, 0x55, 0x6e, 0x69, 0x78, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x43,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x0a, 0x75, 0x6e, 0x69,
	0x78, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x3a, 0x0a, 0x03, 0x74, 0x63, 0x70, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2e, 0x67,
	0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x2e, 0x54,
	0x43, 0x50, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x03,
	0x74, 0x63, 0x70, 0x12, 0x31, 0x0a, 0x14, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x65,
	0x64, 0x5f, 0x63, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x13, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x65, 0x64, 0x43, 0x61, 0x6c,
	0x6c, 0x62, 0x61, 0x63, 0x6b, 0x73, 0x42, 0x0c, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x21, 0x0a, 0x0b, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x22, 0x22, 0x0a, 0x0c, 0x50, 0x69, 0x6e, 0x67, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x2a, 0x54, 0x0a, 0x09, 0x47,
	0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x15, 0x0a, 0x11, 0x47, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12,
	0x19, 0x0a, 0x15, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x52, 0x65, 0x73,
	0x74, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x10, 0x01, 0x12, 0x15, 0x0a, 0x11, 0x47, 0x65,
	0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x4f, 0x70, 0x65, 0x6e, 0x41, 0x50, 0x49, 0x10,
	0x02, 0x32, 0x68, 0x0a, 0x11, 0x52, 0x65, 0x73, 0x74, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x53, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x24,
	0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2e, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x2e, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70, 0x69, 0x2e, 0x67,
	0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x2e, 0x50,
	0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x32, 0x5a, 0x30, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x65, 0x73, 0x68, 0x61, 0x70,
	0x69, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2d, 0x72, 0x65, 0x73, 0x74, 0x2d, 0x67, 0x61, 0x74, 0x65,
	0x77, 0x61, 0x79, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_meshapi_gateway_codegen_plugin_proto_rawDescOnce sync.Once
	file_meshapi_gateway_codegen_plugin_proto_rawDescData = file_meshapi_gateway_codegen_plugin_proto_rawDesc
)

func file_meshapi_gateway_codegen_plugin_proto_rawDescGZIP() []byte {
	file_meshapi_gateway_codegen_plugin_proto_rawDescOnce.Do(func() {
		file_meshapi_gateway_codegen_plugin_proto_rawDescData = protoimpl.X.CompressGZIP(file_meshapi_gateway_codegen_plugin_proto_rawDescData)
	})
	return file_meshapi_gateway_codegen_plugin_proto_rawDescData
}

var file_meshapi_gateway_codegen_plugin_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_meshapi_gateway_codegen_plugin_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_meshapi_gateway_codegen_plugin_proto_goTypes = []interface{}{
	(Generator)(0),               // 0: meshapi.gateway.codegen.Generator
	(*UnixSocketConnection)(nil), // 1: meshapi.gateway.codegen.UnixSocketConnection
	(*TCPConnection)(nil),        // 2: meshapi.gateway.codegen.TCPConnection
	(*Version)(nil),              // 3: meshapi.gateway.codegen.Version
	(*GeneratorInfo)(nil),        // 4: meshapi.gateway.codegen.GeneratorInfo
	(*PluginInfo)(nil),           // 5: meshapi.gateway.codegen.PluginInfo
	(*PingRequest)(nil),          // 6: meshapi.gateway.codegen.PingRequest
	(*PingResponse)(nil),         // 7: meshapi.gateway.codegen.PingResponse
}
var file_meshapi_gateway_codegen_plugin_proto_depIdxs = []int32{
	3, // 0: meshapi.gateway.codegen.GeneratorInfo.version:type_name -> meshapi.gateway.codegen.Version
	0, // 1: meshapi.gateway.codegen.GeneratorInfo.generator:type_name -> meshapi.gateway.codegen.Generator
	1, // 2: meshapi.gateway.codegen.PluginInfo.unix_socket:type_name -> meshapi.gateway.codegen.UnixSocketConnection
	2, // 3: meshapi.gateway.codegen.PluginInfo.tcp:type_name -> meshapi.gateway.codegen.TCPConnection
	6, // 4: meshapi.gateway.codegen.RestGatewayPlugin.Ping:input_type -> meshapi.gateway.codegen.PingRequest
	7, // 5: meshapi.gateway.codegen.RestGatewayPlugin.Ping:output_type -> meshapi.gateway.codegen.PingResponse
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_meshapi_gateway_codegen_plugin_proto_init() }
func file_meshapi_gateway_codegen_plugin_proto_init() {
	if File_meshapi_gateway_codegen_plugin_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_meshapi_gateway_codegen_plugin_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UnixSocketConnection); i {
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
		file_meshapi_gateway_codegen_plugin_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TCPConnection); i {
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
		file_meshapi_gateway_codegen_plugin_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Version); i {
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
		file_meshapi_gateway_codegen_plugin_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GeneratorInfo); i {
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
		file_meshapi_gateway_codegen_plugin_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PluginInfo); i {
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
		file_meshapi_gateway_codegen_plugin_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingRequest); i {
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
		file_meshapi_gateway_codegen_plugin_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingResponse); i {
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
	file_meshapi_gateway_codegen_plugin_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*PluginInfo_UnixSocket)(nil),
		(*PluginInfo_Tcp)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_meshapi_gateway_codegen_plugin_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_meshapi_gateway_codegen_plugin_proto_goTypes,
		DependencyIndexes: file_meshapi_gateway_codegen_plugin_proto_depIdxs,
		EnumInfos:         file_meshapi_gateway_codegen_plugin_proto_enumTypes,
		MessageInfos:      file_meshapi_gateway_codegen_plugin_proto_msgTypes,
	}.Build()
	File_meshapi_gateway_codegen_plugin_proto = out.File
	file_meshapi_gateway_codegen_plugin_proto_rawDesc = nil
	file_meshapi_gateway_codegen_plugin_proto_goTypes = nil
	file_meshapi_gateway_codegen_plugin_proto_depIdxs = nil
}
