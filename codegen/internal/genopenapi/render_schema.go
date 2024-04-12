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
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/internal"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/openapimap"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type fieldSchemaCustomization struct {
	Schema        *openapiv3.Schema
	PathParamName *string
	Required      bool
	ReadOnly      bool
	WriteOnly     bool
}

func (g *Generator) mergeObjects(base, source any) error {
	if g.MergeWithOverwrite {
		if err := mergo.Merge(base, source); err != nil {
			return fmt.Errorf("failed to merge: %w", err)
		}
	} else if err := mergo.Merge(base, source, mergo.WithAppendSlice); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (g *Generator) getCustomizedFieldSchema(
	field *descriptor.Field, config *internal.OpenAPIMessageSpec) (fieldSchemaCustomization, error) {
	result := fieldSchemaCustomization{}

	if config != nil && config.Fields != nil {
		if fieldSchema := config.Fields[field.GetName()]; fieldSchema != nil {
			schemaFromConfig, err := openapimap.Schema(fieldSchema)
			if err != nil {
				return result, fmt.Errorf("failed to map field schema from %q: %w", config.Filename, err)
			}

			result.Schema = schemaFromConfig
			if fieldSchema.Config != nil {
				result.Required = fieldSchema.Config.Required
				if fieldSchema.Config.PathParamName != "" {
					result.PathParamName = &fieldSchema.Config.PathParamName
				}
			}
		}
	}

	if protoSchema, ok := proto.GetExtension(field.Options, api.E_OpenapiField).(*openapi.Schema); ok && protoSchema != nil {
		schemaFromProto, err := openapimap.Schema(protoSchema)
		if err != nil {
			return result, fmt.Errorf("failed to map field schema from %q: %w", field.Message.File.GetName(), err)
		}

		if result.Schema == nil {
			result.Schema = schemaFromProto
		} else {
			if err := g.mergeObjects(result.Schema, schemaFromProto); err != nil {
				return result, err
			}
		}

		if protoSchema.Config != nil {
			if protoSchema.Config.PathParamName != "" && result.PathParamName == nil {
				result.PathParamName = &protoSchema.Config.PathParamName
			}

			if protoSchema.Config.Required {
				result.Required = true
			}
		}
	}

	setFieldAnnotations(field, &result)
	return result, nil
}

// renderComment produces a single description string. This string will NOT be executed with the Go templates yet.
func renderComment(options *Options, location *descriptorpb.SourceCodeInfo_Location) string {
	// get leading comments.
	leadingComments := strings.NewReader(location.GetLeadingComments())
	trailingComments := strings.NewReader(location.GetTrailingComments())

	builder := &strings.Builder{}

	reader := bufio.NewScanner(leadingComments)
	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())
		if options.RemoveInternalComments {
			if strings.HasPrefix(line, "(--") && strings.HasSuffix(line, "--)") {
				continue
			}
		}
		builder.WriteString(line)
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

func (g *Generator) renderEnumComment(enum *descriptor.Enum, values []string) (string, error) {
	comments := g.commentRegistry.LookupEnum(enum)
	if comments == nil {
		return "", nil
	}

	result := renderComment(&g.Options, comments.Location) + "\n"
	for index, value := range values {
		result += "- " + value + ": " + renderComment(&g.Options, comments.Values[int32(index)])
	}

	// TODO: handle the template doc.

	return result, nil
}

func (g *Generator) renderEnumSchema(enum *descriptor.Enum) (*openapiv3.Schema, error) {
	enumConfig := g.enums[enum.FQEN()]
	schema, err := g.getCustomizedEnumSchema(enum, enumConfig)
	if err != nil {
		return nil, err
	}

	if g.UseEnumNumbers {
		values := make([]string, len(enum.Value))
		hasDefault := false
		for index, evdp := range enum.Value {
			values[index] = strconv.FormatInt(int64(evdp.GetNumber()), 10)
			if evdp.GetNumber() == 0 {
				hasDefault = true
			}
		}

		description, err := g.renderEnumComment(enum, values)
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

		if !g.OmitEnumDefaultValue && hasDefault {
			generatedSchema.Object.Default = 0
		}

		if schema != nil {
			if err := g.mergeObjects(schema, generatedSchema); err != nil {
				return nil, err
			}

			return schema, nil
		}

		return generatedSchema, nil
	}

	values := make([]string, len(enum.Value))
	defaultIndex := -1
	for index, evdp := range enum.Value {
		values[index] = evdp.GetName()
		if evdp.GetNumber() == 0 {
			defaultIndex = index
		}
	}

	description, err := g.renderEnumComment(enum, values)
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

	if !g.OmitEnumDefaultValue && defaultIndex != -1 {
		generatedSchema.Object.Default = values[defaultIndex]
	}

	if schema != nil {
		if err := g.mergeObjects(schema, generatedSchema); err != nil {
			return nil, err
		}

		return schema, nil
	}

	return generatedSchema, nil
}

func (g *Generator) createSchemaRef(name string) *openapiv3.Schema {
	return &openapiv3.Schema{
		Object: openapiv3.SchemaCore{
			Ref: "#/components/schemas/" + name,
		},
	}
}

// schemaNameForFQN returns OpenAPI schema name for any enum or message proto FQN.
func (g *Generator) schemaNameForFQN(fqn string) (string, error) {
	result, ok := g.schemaNames[fqn]
	if !ok {
		return "", fmt.Errorf("failed to find OpenAPI name for %q", fqn)
	}

	return result, nil
}

// renderFieldSchema returns OpenAPI schema for a message field.
func (g *Generator) renderFieldSchema(
	field *descriptor.Field,
	baseConfig *openapiv3.Schema) (*openapiv3.Schema, internal.SchemaDependency, error) {

	var fieldSchema *openapiv3.Schema
	dependency := internal.SchemaDependency{}
	repeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED

	switch field.GetType() {
	// TODO: handle the group wire format.
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if wellKnownSchema := wellKnownTypes(field.GetTypeName()); wellKnownSchema != nil {
			fieldSchema = wellKnownSchema
			break
		}
		msg, err := g.registry.LookupMessage(field.Message.File.GetPackage(), field.GetTypeName())
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
			valueSchema, valueDependency, err := g.renderFieldSchema(msg.Fields[1], nil)
			if err != nil {
				return nil, valueDependency, fmt.Errorf("failed to process map entry %q: %w", msg.FQMN(), err)
			}
			repeated = false
			dependency = valueDependency
			fieldSchema.Object.AdditionalProperties = valueSchema
		} else {
			schemaName, err := g.schemaNameForFQN(msg.FQMN())
			if err != nil {
				return nil, dependency, err
			}
			fieldSchema = g.createSchemaRef(schemaName)
			dependency.FQN = msg.FQMN()
			dependency.Kind = internal.DependencyKindMessage
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		enum, err := g.registry.LookupEnum(field.Message.File.GetPackage(), field.GetTypeName())
		if err != nil {
			return nil, dependency, fmt.Errorf("failed to resolve enum %q: %w", field.GetTypeName(), err)
		}
		schemaName, err := g.schemaNameForFQN(enum.FQEN())
		if err != nil {
			return nil, dependency, err
		}
		fieldSchema = g.createSchemaRef(schemaName)
		dependency.FQN = enum.FQEN()
		dependency.Kind = internal.DependencyKindEnum
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
		if err := g.mergeObjects(baseConfig, fieldSchema); err != nil {
			return nil, dependency, err
		}

		return baseConfig, dependency, nil
	}

	return fieldSchema, dependency, nil
}

func setFieldAnnotations(field *descriptor.Field, customizationObject *fieldSchemaCustomization) {
	items, ok := proto.GetExtension(field.Options, annotations.E_FieldBehavior).([]annotations.FieldBehavior)
	if !ok || len(items) == 0 {
		return
	}

	for _, item := range items {
		switch item {
		case annotations.FieldBehavior_REQUIRED:
			customizationObject.Required = true
		case annotations.FieldBehavior_OUTPUT_ONLY:
			customizationObject.ReadOnly = true
		case annotations.FieldBehavior_INPUT_ONLY:
			customizationObject.WriteOnly = true
		}
	}
}

// getCustomizedMessageSchema looks up config files and the proto extensions to prepare the customized OpenAPI v3
// schema for a proto message.
func (g *Generator) getCustomizedMessageSchema(
	message *descriptor.Message, config *internal.OpenAPIMessageSpec) (*openapiv3.Schema, error) {
	var schema *openapiv3.Schema

	if config != nil {
		schemaFromConfig, err := openapimap.Schema(config.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to map message schema from %q: %w", config.Filename, err)
		}

		schema = schemaFromConfig
	}

	if protoSchema, ok := proto.GetExtension(message.Options, api.E_OpenapiSchema).(*openapi.Schema); ok && protoSchema != nil {
		schemaFromProto, err := openapimap.Schema(protoSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to map message schema from %q: %w", message.File.GetName(), err)
		}

		if schema == nil {
			schema = schemaFromProto
		} else {
			if err := g.mergeObjects(schema, schemaFromProto); err != nil {
				return nil, err
			}
		}
	}

	return schema, nil
}

// getCustomizedEnumSchema looks up config files and the proto extensions to prepare the customized OpenAPI v3
// schema for a proto message.
func (g *Generator) getCustomizedEnumSchema(enum *descriptor.Enum, config *internal.OpenAPIEnumSpec) (*openapiv3.Schema, error) {
	var schema *openapiv3.Schema

	if config != nil {
		schemaFromConfig, err := openapimap.Schema(config.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to map enum schema from %q: %w", config.Filename, err)
		}

		schema = schemaFromConfig
	}

	if protoSchema, ok := proto.GetExtension(enum.Options, api.E_OpenapiEnum).(*openapi.Schema); ok && protoSchema != nil {
		schemaFromProto, err := openapimap.Schema(protoSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to map enum schema from %q: %w", enum.File.GetName(), err)
		}

		if schema == nil {
			schema = schemaFromProto
		} else {
			if err := g.mergeObjects(schema, schemaFromProto); err != nil {
				return nil, err
			}
		}
	}

	return schema, nil
}

func (g *Generator) renderMessageSchema(message *descriptor.Message) (internal.OpenAPISchema, error) {
	messageConfig := g.messages[message.FQMN()]
	schema, err := g.getCustomizedMessageSchema(message, messageConfig)
	if err != nil {
		return internal.OpenAPISchema{}, err
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

	var dependencies internal.SchemaDependencyStore

	schema.Object.Properties = make(map[string]*openapiv3.Schema)

	// handle the title, description and summary here.
	comment := g.commentRegistry.LookupMessage(message)
	if comment != nil {
		schema.Object.Description = renderComment(&g.Options, comment.Location)
	}

	for index, field := range message.Fields {
		customFieldSchema, err := g.getCustomizedFieldSchema(field, messageConfig)
		if err != nil {
			return internal.OpenAPISchema{}, fmt.Errorf("failed to parse config for field %q: %w", field.FQFN(), err)
		}

		fieldSchema, dependency, err := g.renderFieldSchema(field, customFieldSchema.Schema)
		if err != nil {
			return internal.OpenAPISchema{}, fmt.Errorf(
				"failed to process field %q in message %q: %w", field.GetName(), message.FQMN(), err)
		}

		if dependency.IsSet() {
			dependencies = internal.AddDependency(dependencies, dependency)
		}

		if fieldSchema.Object.Description == "" && comment != nil && comment.Fields != nil {
			fieldSchema.Object.Description = renderComment(&g.Options, comment.Fields[int32(index)])
		}

		if customFieldSchema.WriteOnly {
			fieldSchema.Object.WriteOnly = true
		}

		if customFieldSchema.ReadOnly {
			fieldSchema.Object.ReadOnly = true
		}

		fieldName := g.fieldName(field)
		schema.Object.Properties[fieldName] = fieldSchema
		if customFieldSchema.Required {
			schema.Object.Required = append(schema.Object.Required, fieldName)
		}
	}

	return internal.OpenAPISchema{Schema: schema, Dependencies: dependencies}, nil
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
