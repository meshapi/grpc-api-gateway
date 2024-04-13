package genopenapi

import (
	"fmt"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/pathfilter"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"github.com/meshapi/grpc-rest-gateway/pkg/httprule"
	"google.golang.org/protobuf/types/descriptorpb"
)

func (s *Session) renderOperation(binding *descriptor.Binding) (*openapiv3.OperationCore, error) {
	operation := &openapiv3.OperationCore{}

	if !s.DisableServiceTags {
		tag := binding.Method.Service.GetName()
		if s.IncludePackageInTags {
			if pkg := binding.Method.Service.File.GetPackage(); pkg != "" {
				tag = pkg + "." + tag
			}
		}

		operation.Tags = append(operation.Tags, tag)
	}

	// handle path parameters
	for _, pathParam := range binding.PathParameters {
		parameter, err := s.renderPathParameter(&pathParam)
		if err != nil {
			return nil, fmt.Errorf("failed to render path parameter %q: %w", pathParam.FieldPath.String(), err)
		}

		operation.Parameters = append(operation.Parameters, &openapiv3.Ref[openapiv3.Parameter]{
			Data: *parameter,
		})
	}

	// handle query parameters
	for _, queryParam := range binding.QueryParameters {
		parameter, err := s.renderQueryParameter(&queryParam)
		if err != nil {
			return nil, fmt.Errorf("failed to render query parameter %q: %w", queryParam.Name, err)
		}

		operation.Parameters = append(operation.Parameters, &openapiv3.Ref[openapiv3.Parameter]{
			Data: *parameter,
		})
	}

	if binding.Body != nil {
		requestBody, err := s.renderRequestBody(binding)
		if err != nil {
			return nil, err
		}

		operation.RequestBody = requestBody
	}

	return operation, nil
}

func (s *Session) renderRequestBody(binding *descriptor.Binding) (*openapiv3.Ref[openapiv3.RequestBody], error) {
	var bodyFilter *pathfilter.Instance
	if len(binding.PathParameters) > 0 {
		bodyFilter = pathfilter.New()
		for _, param := range binding.PathParameters {
			bodyFilter.PutString(param.FieldPath.String())
		}

		if len(binding.Body.FieldPath) > 0 {
			_, bodyFilter = bodyFilter.HasString(binding.Body.FieldPath.String())
		}
	}

	// pull the description, though we likely want to render these operations in a for loop so we can use the same
	// description.

	// we also might want to use the session here so we don't have to add extra values in the slice.
	// TODO: if the target is not a message/group type then we just check exclusion and render field.

	// if the target is indeed a message/group, then we look it up and move on.

	var schema *openapiv3.Schema
	if bodyFilter != nil {
		// render schema with filter.
		requestBody := binding.Method.RequestType
		if len(binding.Body.FieldPath) > 0 {
			fieldPathMessageType := binding.Body.FieldPath.Target().GetTypeName()
			nestedBody, err := s.registry.LookupMessage(requestBody.FQMN(), fieldPathMessageType)
			if err != nil {
				return nil, fmt.Errorf("failed to look up %q: %w", fieldPathMessageType, err)
			}
			requestBody = nestedBody
		}
		filteredSchema, err := s.renderMessageSchemaWithFilter(requestBody, bodyFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to render filtered schema %q: %w", requestBody.FQMN(), err)
		}
		schema = filteredSchema.Schema
		if err := s.includeDependencies(requestBody.FQMN(), filteredSchema.Dependencies); err != nil {
			return nil, err
		}
	} else {
		schemaName, err := s.schemaNameForFQN(binding.Method.RequestType.FQMN())
		if err != nil {
			return nil, fmt.Errorf(
				"could not find schema name for %q: %w", binding.Method.RequestType.FQMN(), err)
		}
		schema = s.createSchemaRef(schemaName)
		if err := s.includeMessage("", binding.Method.RequestType.FQMN()); err != nil {
			return nil, err
		}
	}

	return &openapiv3.Ref[openapiv3.RequestBody]{
		Data: openapiv3.RequestBody{
			Object: openapiv3.RequestBodyCore{
				Content: map[string]*openapiv3.MediaType{
					"application/json": {
						Object: openapiv3.MediaTypeCore{
							Schema: schema,
						},
					},
				},
				Required: true,
			},
		},
	}, nil
}

func (s *Session) renderPathParameter(param *descriptor.Parameter) (*openapiv3.Parameter, error) {

	field := param.Target
	repeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	var schema *openapiv3.Schema

	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if descriptor.IsWellKnownType(field.GetTypeName()) {
			if repeated {
				return nil, fmt.Errorf("only primitive and enum types can be used in repeated path parameters")
			}
			schema = wellKnownTypes(field.GetTypeName())
			break
		}
		return nil, fmt.Errorf("only well-known and primitive types are allowed in path parameters")
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		enum, err := s.registry.LookupEnum(field.Message.File.GetPackage(), field.GetTypeName())
		if err != nil {
			return nil, fmt.Errorf("failed to resolve enum %q: %w", field.GetTypeName(), err)
		}
		schemaName, err := s.schemaNameForFQN(enum.FQEN())
		if err != nil {
			return nil, err
		}
		schema = s.createSchemaRef(schemaName)
		if err := s.includeEnum("", enum.FQEN()); err != nil {
			return nil, err
		}
	default:
		fieldType, format := openAPITypeAndFormatForScalarTypes(field.GetType())
		schema = &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{fieldType},
				Format: format,
			},
		}
	}

	fieldCustomization, err := s.getCustomizedFieldSchema(field, s.messages[field.Message.FQMN()])
	if err != nil {
		return nil, fmt.Errorf("failed to build field customization: %w", err)
	}

	var paramName string
	if fieldCustomization.PathParamName != nil {
		paramName = *fieldCustomization.PathParamName
	} else {
		switch s.Options.FieldNameMode {
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

		switch s.Options.RepeatedPathParameterSeparator {
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

	if fieldCustomization.Schema != nil {
		if err := s.mergeObjects(fieldCustomization.Schema, parameter.Object.Schema); err != nil {
			return nil, err
		}

		parameter.Object.Schema = fieldCustomization.Schema
	}

	if comments := s.commentRegistry.LookupField(param.Target); comments != nil {
		parameter.Object.Description = renderComment(&s.Options, comments)
	}

	return parameter, nil
}

func (s *Session) renderQueryParameter(param *descriptor.QueryParameter) (*openapiv3.Parameter, error) {

	field := param.Target()
	repeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED

	var paramName string
	switch s.Options.FieldNameMode {
	case FieldNameModeJSON:
		if param.NameIsAlias {
			paramName = param.Name
		} else {
			paramName = camelLowerCaseFieldPath(param.FieldPath)
		}
	case FieldNameModeProto:
		paramName = param.Name
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
				return nil, fmt.Errorf("only primitive and enum types can be used in repeated path parameters")
			}
			parameter.Object.Schema = wellKnownTypes(field.GetTypeName())
			break
		}
		message, err := s.registry.LookupMessage(field.Message.FQMN(), field.GetTypeName())
		if err != nil {
			return nil, fmt.Errorf("failed to find message %q: %w", field.GetTypeName(), err)
		}
		if message.IsMapEntry() {
			// NOTE: If the map entry has any type other than scalar types, it cannot be parsed in query parameters.
			if !message.Fields[1].IsScalarType() {
				return nil, fmt.Errorf(
					"only primitive and enum types can be used as map values, received type %q", message.Fields[1].GetTypeName())
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
			return nil, fmt.Errorf("only well-known and primitive types are allowed in path parameters")
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		enum, err := s.registry.LookupEnum(field.Message.File.GetPackage(), field.GetTypeName())
		if err != nil {
			return nil, fmt.Errorf("failed to resolve enum %q: %w", field.GetTypeName(), err)
		}
		schemaName, err := s.schemaNameForFQN(enum.FQEN())
		if err != nil {
			return nil, err
		}
		parameter.Object.Schema = s.createSchemaRef(schemaName)
		if err := s.includeEnum("", enum.FQEN()); err != nil {
			return nil, err
		}
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

		switch s.Options.RepeatedPathParameterSeparator {
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

	fieldCustomization, err := s.getCustomizedFieldSchema(field, s.messages[field.Message.FQMN()])
	if err != nil {
		return nil, fmt.Errorf("failed to build field customization: %w", err)
	}
	if fieldCustomization.Schema != nil {
		if err := s.mergeObjects(fieldCustomization.Schema, parameter.Object.Schema); err != nil {
			return nil, err
		}

		parameter.Object.Schema = fieldCustomization.Schema
	}

	if fieldCustomization.Required {
		parameter.Object.Required = true
	}

	if comments := s.commentRegistry.LookupField(field); comments != nil {
		parameter.Object.Description = renderComment(&s.Options, comments)
	}

	return parameter, nil
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
