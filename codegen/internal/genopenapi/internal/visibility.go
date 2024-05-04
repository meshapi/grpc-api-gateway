package internal

import (
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"google.golang.org/genproto/googleapis/api/visibility"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func GetServiceVisibilityRule(service *descriptor.Service) *visibility.VisibilityRule {
	if service.Options == nil {
		return nil
	}

	if !proto.HasExtension(service.Options, visibility.E_ApiVisibility) {
		return nil
	}

	opts, ok := proto.GetExtension(service.Options, visibility.E_ApiVisibility).(*visibility.VisibilityRule)
	if !ok {
		return nil
	}

	return opts
}

func GetFieldVisibilityRule(field *descriptor.Field) *visibility.VisibilityRule {
	if field.Options == nil {
		return nil
	}

	if !proto.HasExtension(field.Options, visibility.E_FieldVisibility) {
		return nil
	}

	opts, ok := proto.GetExtension(field.Options, visibility.E_FieldVisibility).(*visibility.VisibilityRule)
	if !ok {
		return nil
	}

	return opts
}

func GetMethodVisibilityRule(method *descriptor.Method) *visibility.VisibilityRule {
	if method.Options == nil {
		return nil
	}

	if !proto.HasExtension(method.Options, visibility.E_MethodVisibility) {
		return nil
	}

	opts, ok := proto.GetExtension(method.Options, visibility.E_MethodVisibility).(*visibility.VisibilityRule)
	if !ok {
		return nil
	}

	return opts
}

func GetEnumVisibilityRule(value *descriptorpb.EnumValueDescriptorProto) *visibility.VisibilityRule {
	if value.Options == nil {
		return nil
	}

	if !proto.HasExtension(value.Options, visibility.E_ValueVisibility) {
		return nil
	}

	opts, ok := proto.GetExtension(value.Options, visibility.E_ValueVisibility).(*visibility.VisibilityRule)
	if !ok {
		return nil
	}

	return opts
}
