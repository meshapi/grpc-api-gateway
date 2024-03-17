package genopenapi

import (
	"fmt"
	"strconv"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"google.golang.org/protobuf/types/descriptorpb"
)

func (r *Registry) renderEnumSchema(enum *descriptor.Enum) (*openapiv3.Schema, error) {
	if r.options.UseEnumNumbers {
		values := make([]string, len(enum.Value))
		for index, evdp := range enum.Value {
			values[index] = strconv.FormatInt(int64(evdp.GetNumber()), 10)
		}
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type: openapiv3.TypeSet{openapiv3.TypeInteger},
				Enum: values,
			},
		}, nil
	}

	values := make([]string, len(enum.Value))
	for index, evdp := range enum.Value {
		values[index] = evdp.GetName()
	}
	return &openapiv3.Schema{
		Object: openapiv3.SchemaCore{
			Type: openapiv3.TypeSet{openapiv3.TypeString},
			Enum: values,
		},
	}, nil
}

func (r *Registry) createSchemaRef(name string) *openapiv3.Schema {
	return &openapiv3.Schema{
		Object: openapiv3.SchemaCore{
			Ref: "#/components/schemas/" + name,
		},
	}
}

// schemaNameForFQN returns OpenAPI schema name for any enum or message proto FQN.
func (r *Registry) schemaNameForFQN(fqn string) (string, error) {
	result, ok := r.messageNames[fqn]
	if !ok {
		return "", fmt.Errorf("failed to find OpenAPI name for %q", fqn)
	}

	return result, nil
}

func (r *Registry) renderMessageSchema(message *descriptor.Message) (openAPISchemaConfig, error) {
	schema := &openapiv3.Schema{
		Object: openapiv3.SchemaCore{
			Type: openapiv3.TypeSet{openapiv3.TypeObject},
		},
	}

	var dependencies []string

	schema.Object.Properties = make(map[string]*openapiv3.Schema)

	for _, field := range message.DescriptorProto.Field {
		var fieldSchema *openapiv3.Schema

		repeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED

		switch field.GetType() {
		case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
			// TODO: handle the group wire format.
			//fieldSchema.Object = openapiv3.SchemaCore{
			//  Type: openapiv3.TypeSet{openapiv3.TypeObject},
			//}
		case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
			msg, err := r.descriptorRegistry.LookupMessage(message.File.GetPackage(), field.GetTypeName())
			if err != nil {
				return openAPISchemaConfig{}, fmt.Errorf("failed to resolve message %q: %w", field.GetTypeName(), err)
			}
			schemaName, err := r.schemaNameForFQN(msg.FQMN())
			if err != nil {
				return openAPISchemaConfig{}, err
			}
			fieldSchema = r.createSchemaRef(schemaName)
		case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
			enum, err := r.descriptorRegistry.LookupEnum(message.File.GetPackage(), field.GetTypeName())
			if err != nil {
				return openAPISchemaConfig{}, fmt.Errorf("failed to resolve enum %q: %w", field.GetTypeName(), err)
			}
			schemaName, err := r.schemaNameForFQN(enum.FQEN())
			if err != nil {
				return openAPISchemaConfig{}, err
			}
			fieldSchema = r.createSchemaRef(schemaName)
			dependencies = append(dependencies, field.GetTypeName())
		default:
			fieldType, format := openAPITypeAndFormatForScalarTypes(field.GetType())
			fieldSchema = &openapiv3.Schema{
				Object: openapiv3.SchemaCore{
					Type:   openapiv3.TypeSet{fieldType},
					Format: format,
				},
			}
		}

		if repeated {
			fieldSchema = &openapiv3.Schema{
				Object: openapiv3.SchemaCore{
					Type: openapiv3.TypeSet{openapiv3.TypeArray},
					Items: &openapiv3.ItemSpec{
						Schema: fieldSchema,
					},
				},
			}
		}

		switch r.options.FieldNameMode {
		case FieldNameModeJSON:
			schema.Object.Properties[field.GetJsonName()] = fieldSchema
		case FieldNameModeProto:
			schema.Object.Properties[field.GetName()] = fieldSchema
		}
	}

	// handle the title, description and summary here.

	return openAPISchemaConfig{Schema: schema, Dependencies: dependencies}, nil
}

//func (r *Registry) openAPISchemaForEnum(typeName string) (*openapiv3.SchemaCore, error) {
//  return &openapiv3.SchemaCore{
//    Title: typeName,
//    Type:  openapiv3.TypeSet{openapiv3.TypeString},
//  }
//}

func openAPITypeAndFormatForScalarTypes(t descriptorpb.FieldDescriptorProto_Type) (string, string) {
	switch t {
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return openapiv3.TypeString, ""
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return openapiv3.TypeBoolean, ""
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return openapiv3.TypeNumber, "float"
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return openapiv3.TypeNumber, "double"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return openapiv3.TypeInteger, "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return openapiv3.TypeInteger, "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return openapiv3.TypeInteger, "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return openapiv3.TypeInteger, "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return openapiv3.TypeString, "byte"
	}

	return "", ""
}
