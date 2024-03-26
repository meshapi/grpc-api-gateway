package genopenapi

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"dario.cat/mergo"
	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func (r *Registry) mergeObjects(base, source any) error {
	if r.options.MergeWithOverwrite {
		if err := mergo.Merge(base, source); err != nil {
			return fmt.Errorf("failed to merge: %w", err)
		}
	} else if err := mergo.Merge(base, source, mergo.WithAppendSlice); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

// renderComment produces a single description string. This string will NOT be executed with the Go templates yet.
func (r *Registry) renderComment(location *descriptorpb.SourceCodeInfo_Location) string {
	// get leading comments.
	leadingComments := strings.NewReader(location.GetLeadingComments())
	trailingComments := strings.NewReader(location.GetTrailingComments())

	builder := &strings.Builder{}

	reader := bufio.NewScanner(leadingComments)
	for reader.Scan() {
		builder.WriteString(strings.TrimSpace(reader.Text()))
		builder.WriteByte('\n')
	}

	if trailingComments.Len() > 0 {
		builder.WriteByte('\n')
		reader = bufio.NewScanner(trailingComments)
		for reader.Scan() {
			builder.WriteString(strings.TrimSpace(reader.Text()))
			builder.WriteByte('\n')
		}
	}

	return builder.String()
}

func (r *Registry) renderEnumComment(enum *descriptor.Enum, values []string) (string, error) {
	comments := r.commentRegistry.LookupEnum(enum)
	if comments == nil {
		return "", nil
	}

	result := r.renderComment(comments.Location) + "\n"
	for index, value := range values {
		result += "- " + value + ": " + r.renderComment(comments.Values[int32(index)])
	}

	// TODO: handle the template doc.

	return result, nil
}

func (r *Registry) renderEnumSchema(enum *descriptor.Enum) (*openapiv3.Schema, error) {
	enumConfig := r.enums[enum.FQEN()]
	schema, err := r.getCustomizedEnumSchema(enum, enumConfig)
	if err != nil {
		return nil, err
	}

	if r.options.UseEnumNumbers {
		values := make([]string, len(enum.Value))
		for index, evdp := range enum.Value {
			values[index] = strconv.FormatInt(int64(evdp.GetNumber()), 10)
		}

		description, err := r.renderEnumComment(enum, values)
		if err != nil {
			return nil, fmt.Errorf("failed to process comments: %w", err)
		}

		generatedSchema := &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:        openapiv3.TypeSet{openapiv3.TypeInteger},
				Enum:        values,
				Title:       enum.GetName(),
				Description: description,
			},
		}

		if schema != nil {
			if err := r.mergeObjects(schema, generatedSchema); err != nil {
				return nil, err
			}

			return schema, nil
		}

		return generatedSchema, nil
	}

	values := make([]string, len(enum.Value))
	for index, evdp := range enum.Value {
		values[index] = evdp.GetName()
	}

	description, err := r.renderEnumComment(enum, values)
	if err != nil {
		return nil, fmt.Errorf("failed to process comments: %w", err)
	}

	generatedSchema := &openapiv3.Schema{
		Object: openapiv3.SchemaCore{
			Type:        openapiv3.TypeSet{openapiv3.TypeString},
			Enum:        values,
			Title:       enum.GetName(),
			Description: description,
		},
	}

	if schema != nil {
		if err := r.mergeObjects(schema, generatedSchema); err != nil {
			return nil, err
		}

		return schema, nil
	}

	return generatedSchema, nil
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

// renderFieldSchema returns OpenAPI schema for a message field.
func (r *Registry) renderFieldSchema(
	field *descriptorpb.FieldDescriptorProto,
	message *descriptor.Message,
	baseConfig *openapiv3.Schema) (*openapiv3.Schema, schemaDependency, error) {

	var fieldSchema *openapiv3.Schema
	dependency := schemaDependency{}
	repeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED

	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		// TODO: handle the group wire format.
		//fieldSchema.Object = openapiv3.SchemaCore{
		//  Type: openapiv3.TypeSet{openapiv3.TypeObject},
		//}
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if wellKnownSchema := wellKnownTypes(field.GetTypeName()); wellKnownSchema != nil {
			fieldSchema = wellKnownSchema
			break
		}
		msg, err := r.descriptorRegistry.LookupMessage(message.File.GetPackage(), field.GetTypeName())
		if err != nil {
			return nil, dependency, fmt.Errorf("failed to resolve message %q: %w", field.GetTypeName(), err)
		}
		if msg.IsMapEntry() {
			fieldSchema = &openapiv3.Schema{
				Object: openapiv3.SchemaCore{
					Type: openapiv3.TypeSet{openapiv3.TypeObject},
					AdditionalProperties: &openapiv3.Schema{
						Object: openapiv3.SchemaCore{Type: openapiv3.TypeSet{openapiv3.TypeString}},
					},
				},
			}
			valueSchema, valueDependency, err := r.renderFieldSchema(msg.GetField()[1], msg, nil)
			if err != nil {
				return nil, valueDependency, fmt.Errorf("failed to process map entry %q: %w", msg.FQMN(), err)
			}
			repeated = false
			dependency = valueDependency
			fieldSchema.Object.AdditionalProperties = valueSchema
		} else {
			schemaName, err := r.schemaNameForFQN(msg.FQMN())
			if err != nil {
				return nil, dependency, err
			}
			fieldSchema = r.createSchemaRef(schemaName)
			dependency.FQN = msg.FQMN()
			dependency.Kind = dependencyKindMessage
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		enum, err := r.descriptorRegistry.LookupEnum(message.File.GetPackage(), field.GetTypeName())
		if err != nil {
			return nil, dependency, fmt.Errorf("failed to resolve enum %q: %w", field.GetTypeName(), err)
		}
		schemaName, err := r.schemaNameForFQN(enum.FQEN())
		if err != nil {
			return nil, dependency, err
		}
		fieldSchema = r.createSchemaRef(schemaName)
		dependency.FQN = enum.FQEN()
		dependency.Kind = dependencyKindEnum
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
	fieldSchema.Object.Deprecated = field.Options.GetDeprecated()

	if baseConfig != nil {
		if err := r.mergeObjects(baseConfig, fieldSchema); err != nil {
			return nil, dependency, err
		}

		return baseConfig, dependency, nil
	}

	return fieldSchema, dependency, nil
}

func (r *Registry) setFieldAnnotations(field *descriptor.Field, schema, parent *openapiv3.Schema) {
	items, ok := proto.GetExtension(field.Options, annotations.E_FieldBehavior).([]annotations.FieldBehavior)
	if !ok || len(items) == 0 {
		return
	}

	for _, item := range items {
		switch item {
		case annotations.FieldBehavior_REQUIRED:
			if parent != nil {
				switch r.options.FieldNameMode {
				case FieldNameModeProto:
					parent.Object.Required = append(parent.Object.Required, field.GetName())
				case FieldNameModeJSON:
					parent.Object.Required = append(parent.Object.Required, field.GetJsonName())
				}
			}
		case annotations.FieldBehavior_OUTPUT_ONLY:
			schema.Object.ReadOnly = true
		case annotations.FieldBehavior_INPUT_ONLY:
			schema.Object.WriteOnly = true
		}
	}
}

func (r *Registry) getCustomizedFieldSchema(field *descriptor.Field, config *openAPIMessageConfig) (*openapiv3.Schema, error) {
	var schema *openapiv3.Schema

	if config != nil && config.Fields != nil {
		if fieldSchema := config.Fields[field.GetName()]; fieldSchema != nil {
			schemaFromConfig, err := mapSchema(fieldSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to map field schema from %q: %w", config.Filename, err)
			}

			schema = schemaFromConfig
		}
	}

	if protoSchema, ok := proto.GetExtension(field.Options, api.E_OpenapiField).(*openapi.Schema); ok && protoSchema != nil {
		schemaFromProto, err := mapSchema(protoSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to map field schema from %q: %w", field.Message.File.GetName(), err)
		}

		if schema == nil {
			schema = schemaFromProto
		} else {
			if err := r.mergeObjects(schema, schemaFromProto); err != nil {
				return nil, err
			}
		}
	}

	return schema, nil
}

// getCustomizedMessageSchema looks up config files and the proto extensions to prepare the customized OpenAPI v3
// schema for a proto message.
func (r *Registry) getCustomizedMessageSchema(message *descriptor.Message, config *openAPIMessageConfig) (*openapiv3.Schema, error) {
	var schema *openapiv3.Schema

	if config != nil {
		schemaFromConfig, err := mapSchema(config.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to map message schema from %q: %w", config.Filename, err)
		}

		schema = schemaFromConfig
	}

	if protoSchema, ok := proto.GetExtension(message.Options, api.E_OpenapiSchema).(*openapi.Schema); ok && protoSchema != nil {
		schemaFromProto, err := mapSchema(protoSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to map message schema from %q: %w", message.File.GetName(), err)
		}

		if schema == nil {
			schema = schemaFromProto
		} else {
			if err := r.mergeObjects(schema, schemaFromProto); err != nil {
				return nil, err
			}
		}
	}

	return schema, nil
}

// getCustomizedEnumSchema looks up config files and the proto extensions to prepare the customized OpenAPI v3
// schema for a proto message.
func (r *Registry) getCustomizedEnumSchema(enum *descriptor.Enum, config *openAPIEnumConfig) (*openapiv3.Schema, error) {
	var schema *openapiv3.Schema

	if config != nil {
		schemaFromConfig, err := mapSchema(config.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to map enum schema from %q: %w", config.Filename, err)
		}

		schema = schemaFromConfig
	}

	if protoSchema, ok := proto.GetExtension(enum.Options, api.E_OpenapiEnum).(*openapi.Schema); ok && protoSchema != nil {
		schemaFromProto, err := mapSchema(protoSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to map enum schema from %q: %w", enum.File.GetName(), err)
		}

		if schema == nil {
			schema = schemaFromProto
		} else {
			if err := r.mergeObjects(schema, schemaFromProto); err != nil {
				return nil, err
			}
		}
	}

	return schema, nil
}

func (r *Registry) renderMessageSchema(message *descriptor.Message) (openAPISchemaConfig, error) {
	messageConfig := r.messages[message.FQMN()]
	schema, err := r.getCustomizedMessageSchema(message, messageConfig)
	if err != nil {
		return openAPISchemaConfig{}, err
	}

	if schema == nil {
		schema = &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Title: message.GetName(),
				Type:  openapiv3.TypeSet{openapiv3.TypeObject},
			},
		}
	} else {
		if schema.Object.Title == "" {
			schema.Object.Title = message.GetName()
		}

		if schema.Object.Type == nil {
			schema.Object.Type = openapiv3.TypeSet{openapiv3.TypeObject}
		}
	}

	var dependencies []schemaDependency

	schema.Object.Properties = make(map[string]*openapiv3.Schema)

	// handle the title, description and summary here.
	comment := r.commentRegistry.LookupMessage(message)
	if comment != nil {
		schema.Object.Description = r.renderComment(comment.Location)
	}

	for index, field := range message.Fields {
		customFieldSchema, err := r.getCustomizedFieldSchema(field, messageConfig)
		if err != nil {
			return openAPISchemaConfig{}, fmt.Errorf("failed to parse config for field %q: %w", field.FQFN(), err)
		}

		fieldSchema, dependency, err := r.renderFieldSchema(field.FieldDescriptorProto, message, customFieldSchema)
		if err != nil {
			return openAPISchemaConfig{}, fmt.Errorf(
				"failed to process field %q in message %q: %w", field.GetName(), message.FQMN(), err)
		}

		if dependency.IsSet() {
			dependencies = append(dependencies, dependency)
		}

		if fieldSchema.Object.Description == "" && comment != nil && comment.Fields != nil {
			fieldSchema.Object.Description = r.renderComment(comment.Fields[int32(index)])
		}

		r.setFieldAnnotations(field, fieldSchema, schema)

		switch r.options.FieldNameMode {
		case FieldNameModeJSON:
			schema.Object.Properties[field.GetJsonName()] = fieldSchema
		case FieldNameModeProto:
			schema.Object.Properties[field.GetName()] = fieldSchema
		}
	}

	return openAPISchemaConfig{Schema: schema, Dependencies: dependencies}, nil
}

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

func wellKnownTypes(fqmn string) *openapiv3.Schema {
	switch fqmn {
	case ".google.protobuf.FieldMask":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type: openapiv3.TypeSet{openapiv3.TypeString},
			},
		}
	case ".google.protobuf.Timestamp":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{openapiv3.TypeString},
				Format: "date-time",
			},
		}
	case ".google.protobuf.Duration":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type: openapiv3.TypeSet{openapiv3.TypeString},
			},
		}
	case ".google.protobuf.StringValue":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type: openapiv3.TypeSet{openapiv3.TypeString},
			},
		}
	case ".google.protobuf.BytesValue":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{openapiv3.TypeString},
				Format: "byte",
			},
		}
	case ".google.protobuf.Int32Value":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{openapiv3.TypeInteger},
				Format: "int32",
			},
		}
	case ".google.protobuf.UInt32Value":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{openapiv3.TypeInteger},
				Format: "uint32",
			},
		}
	case ".google.protobuf.Int64Value":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{openapiv3.TypeInteger},
				Format: "int64",
			},
		}
	case ".google.protobuf.UInt64Value":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{openapiv3.TypeInteger},
				Format: "uint64",
			},
		}
	case ".google.protobuf.FloatValue":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{openapiv3.TypeNumber},
				Format: "float",
			},
		}
	case ".google.protobuf.DoubleValue":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{openapiv3.TypeNumber},
				Format: "double",
			},
		}
	case ".google.protobuf.BoolValue":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type: openapiv3.TypeSet{openapiv3.TypeBoolean},
			},
		}
	case ".google.protobuf.Empty", ".google.protobuf.Struct":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type: openapiv3.TypeSet{openapiv3.TypeObject},
			},
		}
	case ".google.protobuf.Value":
		return &openapiv3.Schema{}
	case ".google.protobuf.ListValue":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type: openapiv3.TypeSet{openapiv3.TypeArray},
				Items: &openapiv3.ItemSpec{
					Schema: &openapiv3.Schema{
						Object: openapiv3.SchemaCore{
							Type: openapiv3.TypeSet{openapiv3.TypeObject},
						},
					},
				},
			},
		}
	case ".google.protobuf.NullValue":
		return &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type: openapiv3.TypeSet{openapiv3.TypeNull},
			},
		}
	}

	return nil
}
