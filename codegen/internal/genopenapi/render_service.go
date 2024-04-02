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
		parameter, dependency, err := g.renderPathParameter(&pathParam, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to render path parameter: %w", err)
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
			return nil, nil, fmt.Errorf("failed to render path parameter: %w", err)
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
	param *descriptor.Parameter,
	baseConfig *openapiv3.Schema) (*openapiv3.Parameter, schemaDependency, error) {

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
			Name:     paramName,
			Schema:   schema,
			In:       openapiv3.ParameterInQuery,
			Required: false, // only if the schema says it is required.
		},
	}

	// TODO:
	// 1. Perhaps have the annotations to return whether or not we have a required field or not.
	// that way, here we can set the query parameter as required and elsewhere handle required array.
	// 2. also make it a function instead of a method.

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
