package genopenapi

import (
	"fmt"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"github.com/meshapi/grpc-rest-gateway/pkg/httprule"
	"google.golang.org/protobuf/types/descriptorpb"
)

func (g *Generator) renderOperation(
	binding *descriptor.Binding) (*openapiv3.OperationCore, []schemaDependency, error) {
	operation := &openapiv3.OperationCore{}

	var dependencies []schemaDependency

	if !g.DisableServiceTags {
		tag := binding.Method.Service.GetName()
		if g.IncludePackageInTags {
			if pkg := binding.Method.Service.File.GetPackage(); pkg != "" {
				tag = pkg + "." + tag
			}
		}

		operation.Tags = append(operation.Tags, tag)
	}

	// handle path parameters
	for _, pathParam := range binding.PathParameters {
		parameter, dependency, err := g.renderPathParameter(&pathParam)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to render path parameter %q: %w", pathParam.FieldPath.String(), err)
		}

		if dependency.IsSet() {
			dependencies = append(dependencies, dependency)
		}

		operation.Parameters = append(operation.Parameters, &openapiv3.Ref[openapiv3.Parameter]{
			Data: *parameter,
		})
	}

	// handle query parameters
	for _, queryParam := range binding.QueryParameters {
		parameter, dependency, err := g.renderQueryParameter(&queryParam, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to render query parameter %q: %w", queryParam.Name, err)
		}
		if parameter == nil { // if parameter is not produced, skip it.
			continue
		}

		if dependency.IsSet() {
			dependencies = append(dependencies, dependency)
		}

		operation.Parameters = append(operation.Parameters, &openapiv3.Ref[openapiv3.Parameter]{
			Data: *parameter,
		})
	}

	return operation, dependencies, nil
}

func (g *Generator) renderPathParameter(
	param *descriptor.Parameter) (*openapiv3.Parameter, schemaDependency, error) {

	field := param.Target
	repeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	var schema *openapiv3.Schema
	var dependency schemaDependency

	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if descriptor.IsWellKnownType(field.GetTypeName()) {
			if repeated {
				return nil, dependency, fmt.Errorf("only primitive and enum types can be used in repeated path parameters")
			}
			schema = wellKnownTypes(field.GetTypeName())
			break
		}
		return nil, dependency, fmt.Errorf("only well-known and primitive types are allowed in path parameters")
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		enum, err := g.registry.LookupEnum(field.Message.File.GetPackage(), field.GetTypeName())
		if err != nil {
			return nil, dependency, fmt.Errorf("failed to resolve enum %q: %w", field.GetTypeName(), err)
		}
		schemaName, err := g.openapiRegistry.schemaNameForFQN(enum.FQEN())
		if err != nil {
			return nil, dependency, err
		}
		schema = g.openapiRegistry.createSchemaRef(schemaName)
		dependency.FQN = enum.FQEN()
		dependency.Kind = dependencyKindEnum
	default:
		fieldType, format := openAPITypeAndFormatForScalarTypes(field.GetType())
		schema = &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{fieldType},
				Format: format,
			},
		}
	}

	var paramName string
	if config := g.openapiRegistry.lookUpFieldConfig(field); config != nil && config.PathParamName != "" {
		paramName = config.PathParamName
	} else {
		switch g.Options.FieldNameMode {
		case FieldNameModeJSON:
			paramName = camelLowerCaseFieldPath(param.FieldPath)
		case FieldNameModeProto:
			paramName = param.String()
		}
	}

	parameter := &openapiv3.Parameter{
		Object: openapiv3.ParameterCore{
			Name:     paramName,
			Schema:   schema,
			In:       openapiv3.ParameterInPath,
			Required: true,
		},
	}

	if repeated {
		// handle the repeated attributes here.
		parameter.Object.Schema = &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:     openapiv3.TypeSet{openapiv3.TypeArray},
				Items:    &openapiv3.ItemSpec{Schema: parameter.Object.Schema},
				MinItems: 1,
			},
		}

		switch g.Options.RepeatedPathParameterSeparator {
		case descriptor.PathParameterSeparatorCSV:
			parameter.Object.Style = openapiv3.ParameterStyleSimple
		case descriptor.PathParameterSeparatorSSV:
			parameter.Object.Style = openapiv3.ParameterStyleSpaceDelimited
		case descriptor.PathParameterSeparatorTSV:
			parameter.Object.Style = openapiv3.ParameterStyleTabDelimited
		case descriptor.PathParameterSeparatorPipes:
			parameter.Object.Style = openapiv3.ParameterStylePipeDelimited
		}
	}

	fieldCustomization, err := g.openapiRegistry.getCustomizedFieldSchema(
		field, g.openapiRegistry.messages[field.Message.FQMN()])
	if err != nil {
		return nil, dependency, fmt.Errorf("failed to build field customization: %w", err)
	}
	if fieldCustomization.Schema != nil {
		if err := g.openapiRegistry.mergeObjects(fieldCustomization.Schema, parameter.Object.Schema); err != nil {
			return nil, dependency, err
		}

		parameter.Object.Schema = fieldCustomization.Schema
	}

	if comments := g.commentRegistry.LookupField(param.Target); comments != nil {
		parameter.Object.Description = renderComment(&g.Options, comments)
	}

	return parameter, dependency, nil
}

func (g *Generator) renderQueryParameter(
	param *descriptor.QueryParameter,
	baseConfig *openapiv3.Schema) (*openapiv3.Parameter, schemaDependency, error) {

	field := param.Target()
	repeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	var dependency schemaDependency

	var paramName string
	if config := g.openapiRegistry.lookUpFieldConfig(field); config != nil && config.PathParamName != "" {
		paramName = config.PathParamName
	} else {
		switch g.Options.FieldNameMode {
		case FieldNameModeJSON:
			if param.NameIsAlias {
				paramName = param.Name
			} else {
				paramName = camelLowerCaseFieldPath(param.FieldPath)
			}
		case FieldNameModeProto:
			paramName = param.Name
		}
	}

	parameter := &openapiv3.Parameter{
		Object: openapiv3.ParameterCore{
			Name:       paramName,
			In:         openapiv3.ParameterInQuery,
			Deprecated: field.GetOptions().GetDeprecated(),
		},
	}

	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if descriptor.IsWellKnownType(field.GetTypeName()) {
			if repeated {
				return nil, dependency, fmt.Errorf("only primitive and enum types can be used in repeated path parameters")
			}
			parameter.Object.Schema = wellKnownTypes(field.GetTypeName())
			break
		}
		message, err := g.registry.LookupMessage(field.Message.FQMN(), field.GetTypeName())
		if err != nil {
			return nil, dependency, fmt.Errorf("failed to find message %q: %w", field.GetTypeName(), err)
		}
		if message.IsMapEntry() {
			// NOTE: If the map entry has any type other than scalar types, it cannot be parsed in query parameters.
			// TODO: We might want to write a warning as an option or return an error if the query is not auto-mapped.
			if !message.Fields[1].IsScalarType() {
				return nil, dependency, nil
			}
			repeated = false
			parameter.Object.Schema = &openapiv3.Schema{
				Object: openapiv3.SchemaCore{
					Type:   openapiv3.TypeSet{openapiv3.TypeString},
					Format: "map",
				},
			}
			parameter.Object.Description = `This is a request variable of the map type.` +
				` The query format is "map_name[key]=value", e.g. If the map name is Age, the key type is string,` +
				` and the value type is integer, the query parameter is expressed as Age["bob"]=18`
		} else {
			return nil, dependency, fmt.Errorf("only well-known and primitive types are allowed in path parameters")
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		enum, err := g.registry.LookupEnum(field.Message.File.GetPackage(), field.GetTypeName())
		if err != nil {
			return nil, dependency, fmt.Errorf("failed to resolve enum %q: %w", field.GetTypeName(), err)
		}
		schemaName, err := g.openapiRegistry.schemaNameForFQN(enum.FQEN())
		if err != nil {
			return nil, dependency, err
		}
		parameter.Object.Schema = g.openapiRegistry.createSchemaRef(schemaName)
		dependency.FQN = enum.FQEN()
		dependency.Kind = dependencyKindEnum
	default:
		fieldType, format := openAPITypeAndFormatForScalarTypes(field.GetType())
		parameter.Object.Schema = &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{fieldType},
				Format: format,
			},
		}
	}

	if repeated {
		// handle the repeated attributes here.
		parameter.Object.Schema = &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:     openapiv3.TypeSet{openapiv3.TypeArray},
				Items:    &openapiv3.ItemSpec{Schema: parameter.Object.Schema},
				MinItems: 1,
			},
		}

		switch g.Options.RepeatedPathParameterSeparator {
		case descriptor.PathParameterSeparatorCSV:
			parameter.Object.Style = openapiv3.ParameterStyleSimple
		case descriptor.PathParameterSeparatorSSV:
			parameter.Object.Style = openapiv3.ParameterStyleSpaceDelimited
		case descriptor.PathParameterSeparatorTSV:
			parameter.Object.Style = openapiv3.ParameterStyleTabDelimited
		case descriptor.PathParameterSeparatorPipes:
			parameter.Object.Style = openapiv3.ParameterStylePipeDelimited
		}
	}

	fieldCustomization, err := g.openapiRegistry.getCustomizedFieldSchema(
		field, g.openapiRegistry.messages[field.Message.FQMN()])
	if err != nil {
		return nil, dependency, fmt.Errorf("failed to build field customization: %w", err)
	}
	if fieldCustomization.Schema != nil {
		if err := g.openapiRegistry.mergeObjects(fieldCustomization.Schema, parameter.Object.Schema); err != nil {
			return nil, dependency, err
		}

		parameter.Object.Schema = fieldCustomization.Schema
	}

	if fieldCustomization.Required {
		parameter.Object.Required = true
	}

	if comments := g.commentRegistry.LookupField(field); comments != nil {
		parameter.Object.Description = renderComment(&g.Options, comments)
	}

	return parameter, dependency, nil
}

func camelLowerCaseFieldPath(fieldPath descriptor.FieldPath) string {
	builder := &strings.Builder{}
	for index, part := range fieldPath {
		if index != 0 {
			builder.WriteByte('.')
		}
		builder.WriteString(part.Target.GetJsonName())
	}

	return builder.String()
}

func renderPath(binding *descriptor.Binding) string {
	writer := &strings.Builder{}

	if len(binding.PathTemplate.Segments) == 0 {
		return "/"
	}

	for _, segment := range binding.PathTemplate.Segments {
		switch segment.Type {
		case httprule.SegmentTypeLiteral:
			writer.WriteString("/" + segment.Value)
		case httprule.SegmentTypeSelector:
			writer.WriteString("/{" + segment.Value + "}")
		case httprule.SegmentTypeWildcard:
			_, _ = fmt.Fprintf(writer, "/?")
		case httprule.SegmentTypeCatchAllSelector:
			writer.WriteString("/{" + segment.Value + "}")
		default:
			_, _ = fmt.Fprintf(writer, "/<!?:%s>", segment.Value)
		}
	}

	return writer.String()
}
