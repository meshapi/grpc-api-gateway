package genopenapi

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"dario.cat/mergo"
	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/internal"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/openapimap"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/protocomment"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/genproto/googleapis/api/visibility"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

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

func (g *Generator) mergeObjectsOverride(base, source any) error {
	if g.MergeWithOverwrite {
		if err := mergo.Merge(base, source, mergo.WithOverride); err != nil {
			return fmt.Errorf("failed to merge: %w", err)
		}
	} else if err := mergo.Merge(base, source, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (g *Generator) getCustomizedFieldSchema(
	field *descriptor.Field, config *internal.OpenAPIMessageSpec) (internal.FieldSchemaCustomization, error) {
	result := internal.FieldSchemaCustomization{}

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

	var protoSchema *openapi.Schema
	var ok bool
	if field.Options != nil && proto.HasExtension(field.Options, api.E_OpenapiField) {
		protoSchema, ok = proto.GetExtension(field.Options, api.E_OpenapiField).(*openapi.Schema)
	}
	if ok && protoSchema != nil {
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
func (g *Generator) renderComment(location *descriptorpb.SourceCodeInfo_Location) string {
	if g.IgnoreComments {
		return ""
	}

	// get leading comments.
	leadingComments := strings.NewReader(location.GetLeadingComments())
	trailingComments := strings.NewReader(location.GetTrailingComments())

	builder := &strings.Builder{}

	reader := bufio.NewScanner(leadingComments)
	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())
		if g.RemoveInternalComments {
			if strings.HasPrefix(line, commentInternalOpen) && strings.HasSuffix(line, commentInternalClose) {
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

	result := g.renderComment(comments.Location) + "\n"
	for index, value := range values {
		result += "- " + value + ": " + g.renderComment(comments.Values[int32(index)])
	}

	return result, nil
}

func (g *Generator) evaluateCommentWithTemplate(body string, data any) string {
	tpl, err := template.New("").Funcs(template.FuncMap{
		"import": func(name string) string {
			file, err := os.ReadFile(name)
			if err != nil {
				return err.Error()
			}

			return g.evaluateCommentWithTemplate(string(file), data)
		},
		"fieldcomments": func(field *descriptor.Field) string {
			return strings.ReplaceAll(g.renderComment(g.commentRegistry.LookupField(field)), "\n", "<br>")
		},
		"arg": func(name string) string {
			for _, v := range g.GoTemplateArgs {
				if v.Key == name {
					return v.Value
				}
			}
			return fmt.Sprintf("Go template argument %q not found", name)
		},
	}).Parse(body)
	if err != nil {
		// If there is an error parsing the templating insert the error as string in the comment
		// to make it easier to debug the template error.
		return err.Error()
	}
	var tmp bytes.Buffer
	if err := tpl.Execute(&tmp, data); err != nil {
		// If there is an error executing the templating insert the error as string in the comment
		// to make it easier to debug the error.
		return err.Error()
	}
	return tmp.String()
}

func (g *Generator) isVisible(v *visibility.VisibilityRule) bool {
	if v == nil || v.Restriction == "" {
		return true
	}

	startIndex := 0
	for {
		index := strings.IndexRune(v.Restriction[startIndex:], ',')
		if index == -1 {
			break
		}
		if g.VisibilitySelectors[strings.TrimSpace(v.Restriction[startIndex:startIndex+index])] {
			return true
		}
		startIndex = startIndex + index + 1
	}
	return g.VisibilitySelectors[strings.TrimSpace(v.Restriction[startIndex:])]
}

func (g *Generator) renderEnumSchema(enum *descriptor.Enum) (*openapiv3.Schema, error) {
	enumConfig := g.enums[enum.FQEN()]
	schema, err := g.getCustomizedEnumSchema(enum, enumConfig)
	if err != nil {
		return nil, err
	}

	if g.UseEnumNumbers {
		values := make([]string, 0, len(enum.Value))
		hasDefault := false
		for _, evdp := range enum.Value {
			if !g.isVisible(internal.GetEnumVisibilityRule(evdp)) {
				continue
			}
			values = append(values, strconv.FormatInt(int64(evdp.GetNumber()), 10))
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

	values := make([]string, 0, len(enum.Value))
	defaultIndex := -1
	for index, evdp := range enum.Value {
		if !g.isVisible(internal.GetEnumVisibilityRule(evdp)) {
			continue
		}

		values = append(values, evdp.GetName())
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

		generatedSchema = schema
	}

	if g.UseGoTemplate && generatedSchema.Object.Description != "" {
		generatedSchema.Object.Description = g.evaluateCommentWithTemplate(generatedSchema.Object.Description, enum)
	}

	return generatedSchema, nil
}

func (g *Generator) createSchemaRef(name string) *openapiv3.Schema {
	return &openapiv3.Schema{
		Object: openapiv3.SchemaCore{
			Ref: refPrefix + name,
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
	customization internal.FieldSchemaCustomization,
	comments *protocomment.Location) (*openapiv3.Schema, internal.SchemaDependency, error) {

	var fieldSchema *openapiv3.Schema
	dependency := internal.SchemaDependency{}
	repeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED

	switch field.GetType() {
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
			valueSchema, valueDependency, err := g.renderFieldSchema(msg.Fields[1], internal.FieldSchemaCustomization{}, nil)
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

	if customization.WriteOnly {
		fieldSchema.Object.WriteOnly = true
	}

	if customization.ReadOnly {
		fieldSchema.Object.ReadOnly = true
	}

	if customization.Schema != nil {
		if err := g.mergeObjects(customization.Schema, fieldSchema); err != nil {
			return nil, dependency, err
		}

		fieldSchema = customization.Schema
	}

	if comments != nil && fieldSchema.Object.Description == "" {
		fieldSchema.Object.Description = g.renderComment(comments)
	}

	if g.UseGoTemplate && fieldSchema.Object.Description != "" {
		fieldSchema.Object.Description = g.evaluateCommentWithTemplate(fieldSchema.Object.Description, field)
	}

	return fieldSchema, dependency, nil
}

func setFieldAnnotations(field *descriptor.Field, customizationObject *internal.FieldSchemaCustomization) {
	if field.Options == nil {
		return
	}
	if !proto.HasExtension(field.Options, annotations.E_FieldBehavior) {
		return
	}
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

	if message.Options == nil || !proto.HasExtension(message.Options, api.E_OpenapiSchema) {
		return schema, nil
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

	if enum.Options == nil || !proto.HasExtension(enum.Options, api.E_OpenapiEnum) {
		return schema, nil
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
	if comment != nil && schema.Object.Description == "" {
		schema.Object.Description = g.renderComment(comment.Location)
	}

	if g.UseGoTemplate && schema.Object.Description != "" {
		schema.Object.Description = g.evaluateCommentWithTemplate(schema.Object.Description, message)
	}

	requiredSet := internal.RequiredSetFromSlice(schema.Object.Required)

	for index, field := range message.Fields {
		if !g.isVisible(internal.GetFieldVisibilityRule(field)) {
			continue
		}

		customFieldSchema, err := g.getCustomizedFieldSchema(field, messageConfig)
		if err != nil {
			return internal.OpenAPISchema{}, fmt.Errorf("failed to parse config for field %q: %w", field.FQFN(), err)
		}

		var fieldComments *protocomment.Location
		if comment != nil && comment.Fields != nil {
			fieldComments = comment.Fields[int32(index)]
		}
		fieldSchema, dependency, err := g.renderFieldSchema(field, customFieldSchema, fieldComments)
		if err != nil {
			return internal.OpenAPISchema{}, fmt.Errorf(
				"failed to process field %q in message %q: %w", field.GetName(), message.FQMN(), err)
		}

		if dependency.IsSet() {
			dependencies = internal.AddDependency(dependencies, dependency)
		}

		fieldName := g.fieldName(field)
		if customFieldSchema.Required || g.fieldIsDeemedRequired(field) {
			requiredSet = internal.AppendToRequiredSet(requiredSet, fieldName)
		}

		if g.FieldNullableMode == FieldNullableModeOptionalLabel {
			if field.GetProto3Optional() {
				g.makeFieldSchemaNullable(fieldSchema)
			}
		} else if g.FieldNullableMode == FieldNullableModeNonRequired {
			if field.GetProto3Optional() {
				g.makeFieldSchemaNullable(fieldSchema)
			} else if _, isRequired := requiredSet[fieldName]; !isRequired {
				g.makeFieldSchemaNullable(fieldSchema)
			}
		}
		schema.Object.Properties[fieldName] = fieldSchema
	}

	schema.Object.Required = internal.RequiredSliceFromRequiredSet(requiredSet)

	return internal.OpenAPISchema{Schema: schema, Dependencies: dependencies}, nil
}

// makeFieldSchemaNullable updates the field schema so that it accepts null as a value.
// if ref is available, oneOf will be used, otherwise, a null type will be appended to the type array unless the type
// array already includes the null value.
func (g *Generator) makeFieldSchemaNullable(schema *openapiv3.Schema) {
	if schema.Object.Ref != "" {
		schema.Object.OneOf = []*openapiv3.Schema{
			{Object: openapiv3.SchemaCore{Ref: schema.Object.Ref}},
			{Object: openapiv3.SchemaCore{Type: openapiv3.TypeSet{openapiv3.TypeNull}}},
		}
		schema.Object.Ref = ""
		return

	}

	for index := range schema.Object.Type {
		if schema.Object.Type[index] == openapiv3.TypeNull {
			return
		}
	}
	schema.Object.Type = append(schema.Object.Type, openapiv3.TypeNull)
}

// fieldIsDeemedRequired uses the field required mode option to determine if the current field should be deemed requried
// or not.
func (g *Generator) fieldIsDeemedRequired(field *descriptor.Field) bool {
	if g.FieldRequiredMode == FieldRequiredModeDisabled {
		return false
	}

	if g.FieldRequiredMode == FieldRequiredModeRequireNonOptionalScalar {
		return !field.GetProto3Optional() && field.IsScalarType()
	}

	return !field.GetProto3Optional()
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
