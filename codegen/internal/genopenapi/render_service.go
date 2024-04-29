package genopenapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/internal"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/openapimap"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/pathfilter"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"github.com/meshapi/grpc-rest-gateway/pkg/httprule"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func (s *Session) renderOperation(
	binding *descriptor.Binding,
	defaultResponses internal.DefaultResponses) (*openapiv3.Operation, error) {
	operation := &openapiv3.Operation{}

	switch s.OperationIDMode {
	case OperationIDModeFQN:
		operation.Object.OperationID = binding.Method.FQMN()[1:] // NB: remove the leading dot.
	case OperationIDModeServiceAndMethod:
		operation.Object.OperationID = binding.Method.Service.GetName() + "_" + binding.Method.GetName()
	case OperationIDModeMethod:
		operation.Object.OperationID = binding.Method.GetName()
	}

	if binding.Index > 0 {
		operation.Object.OperationID += strconv.Itoa(binding.Index + 1)
	}

	if !s.DisableServiceTags {
		operation.Object.Tags = append(operation.Object.Tags, s.tagNameForService(binding.Method.Service))
	}

	// handle path parameters
	for _, pathParam := range binding.PathParameters {
		parameter, err := s.renderPathParameter(&pathParam)
		if err != nil {
			return nil, fmt.Errorf("failed to render path parameter %q: %w", pathParam.FieldPath.String(), err)
		}

		operation.Object.Parameters = append(operation.Object.Parameters, &openapiv3.Ref[openapiv3.Parameter]{
			Data: *parameter,
		})
	}

	// handle query parameters
	for _, queryParam := range binding.QueryParameters {
		if s.AllowPatchFeature && binding.HTTPMethod == http.MethodPatch {
			target := queryParam.Target()
			if target.GetTypeName() == fqmnFieldMask && target.GetName() == fieldNameUpdateMask {
				continue
			}
		}

		parameter, err := s.renderQueryParameter(&queryParam)
		if err != nil {
			return nil, fmt.Errorf("failed to render query parameter %q: %w", queryParam.Name, err)
		}

		operation.Object.Parameters = append(operation.Object.Parameters, &openapiv3.Ref[openapiv3.Parameter]{
			Data: *parameter,
		})
	}

	if binding.Body != nil {
		requestBody, err := s.renderRequestBody(binding)
		if err != nil {
			return nil, err
		}

		operation.Object.RequestBody = requestBody
	}

	if err := s.addDefaultResponses(defaultResponses, &operation.Object); err != nil {
		return nil, err
	}

	if !s.DisableDefaultResponses {
		if err := s.addDefaultSuccessResponse(binding, &operation.Object); err != nil {
			return nil, fmt.Errorf("failed to add default responses: %w", err)
		}
	}

	if err := s.addDefaultErrorResponse(&operation.Object); err != nil {
		return nil, err
	}

	return operation, nil
}

func (s *Session) addDefaultErrorResponse(operation *openapiv3.OperationCore) error {
	if s.DisableDefaultResponses || operation.Responses != nil && operation.Responses[httpStatusDefault] != nil {
		return nil
	}

	response := internal.DefaultErrorResponse()
	if !response.ReferenceIsResolved {
		ref, err := s.schemaNameForFQN(rpcStatusProto)
		if err != nil {
			return fmt.Errorf("unexpected error: %w", err)
		}
		internal.SetErrorResponseRef(refPrefix + ref)
	}

	if operation.Responses == nil {
		operation.Responses = map[string]*openapiv3.Ref[openapiv3.Response]{
			httpStatusDefault: response.Response,
		}
	} else {
		operation.Responses[httpStatusDefault] = response.Response
	}

	if !s.includedDefaultErrorStatusDependency {
		if err := s.includeMessage(rpcStatusProto); err != nil {
			return fmt.Errorf("unexpected error importing rpc.Status: %w", err)
		}
		s.includedDefaultErrorStatusDependency = true
	}

	return nil
}

func (s *Session) addDefaultResponses(responses internal.DefaultResponses, operation *openapiv3.OperationCore) error {
	if len(responses) == 0 {
		return nil
	}

	if operation.Responses == nil {
		operation.Responses = make(map[string]*openapiv3.Ref[openapiv3.Response])
	}

	for status, response := range responses {
		operation.Responses[status] = response.Response
		if len(response.Dependencies) == 0 {
			continue
		}

		if err := s.includeDependencies(response.Dependencies); err != nil {
			return err
		}
	}

	return nil
}

func (s *Session) addFieldSchema(field *descriptor.Field) (*openapiv3.Schema, error) {
	fieldCustomization, err := s.getCustomizedFieldSchema(field, s.messages[field.Message.FQMN()])
	if err != nil {
		return nil, fmt.Errorf("failed to prepare field customization for %q: %w", field.Message.FQMN(), err)
	}
	fieldSchema, dependency, err := s.renderFieldSchema(
		field, fieldCustomization, s.commentRegistry.LookupField(field))
	if err != nil {
		return nil, fmt.Errorf("failed to render field %q: %w", field.FQFN(), err)
	}
	if dependency.IsSet() {
		if err := s.includeDependency(field.Message.FQMN(), dependency); err != nil {
			return nil, fmt.Errorf("failed to render dependency %q: %w", dependency.FQN, err)
		}
	}
	return fieldSchema, nil
}

func (s *Session) renderRequestBody(binding *descriptor.Binding) (*openapiv3.Ref[openapiv3.RequestBody], error) {
	var bodyFilter *pathfilter.Instance
	hasPathParameters := len(binding.PathParameters) > 0
	hasBodySelector := len(binding.Body.FieldPath) > 0
	if hasPathParameters {
		bodyFilter = pathfilter.New()
		for _, param := range binding.PathParameters {
			bodyFilter.PutString(param.FieldPath.String())
		}
	}

	var schema *openapiv3.Schema
	var field *descriptor.Field
	requestBody := binding.Method.RequestType

	if hasBodySelector {
		field = binding.Body.FieldPath.Target()
		if field.IsScalarType() {
			// if request body field path resolves to a scalar type, it will be rendered as a singular field or
			// ignored completely if it is already captured in the path.
			if bodyFilter != nil {
				if filteredOut, _ := bodyFilter.HasString(binding.Body.FieldPath.String()); filteredOut {
					return nil, nil
				}
			}
			fieldSchema, err := s.addFieldSchema(field)
			if err != nil {
				return nil, err
			}
			schema = fieldSchema
		} else {
			_, bodyFilter = bodyFilter.HasString(binding.Body.FieldPath.String())
			nestedBody, err := s.registry.LookupMessage(requestBody.FQMN(), field.GetTypeName())
			if err != nil {
				return nil, fmt.Errorf("failed to look up %q: %w", field.GetTypeName(), err)
			}
			requestBody = nestedBody
		}
	}

	if schema == nil {
		switch {
		case bodyFilter != nil:
			// render schema with filter.
			filteredSchema, err := s.renderMessageSchemaWithFilter(requestBody, bodyFilter)
			if err != nil {
				return nil, fmt.Errorf("failed to render filtered schema %q: %w", requestBody.FQMN(), err)
			}
			schema = filteredSchema.Schema
			if err := s.includeDependencies(filteredSchema.Dependencies); err != nil {
				return nil, err
			}
		case field != nil:
			fieldSchema, err := s.addFieldSchema(field)
			if err != nil {
				return nil, err
			}
			schema = fieldSchema
		default:
			schemaName, err := s.schemaNameForFQN(requestBody.FQMN())
			if err != nil {
				return nil, fmt.Errorf(
					"could not find schema name for %q: %w", requestBody.FQMN(), err)
			}
			schema = s.createSchemaRef(schemaName)
			if err := s.includeMessage(requestBody.FQMN()); err != nil {
				return nil, err
			}
		}
	}

	request := &openapiv3.Ref[openapiv3.RequestBody]{
		Data: openapiv3.RequestBody{
			Object: openapiv3.RequestBodyCore{
				Content: map[string]*openapiv3.MediaType{
					mimeTypeJSON: {
						Object: openapiv3.MediaTypeCore{
							Schema: schema,
						},
					},
				},
				Required: true,
			},
		},
	}

	if binding.Method.GetClientStreaming() {
		request.Data.Object.Description += streamingInputDescription
	}

	return request, nil
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
		if err := s.includeEnum(enum.FQEN()); err != nil {
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
		parameter.Object.Description = s.renderComment(comments)
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
		if descriptor.IsWellKnownType(field.GetTypeName()) || field.GetTypeName() == fqmnFieldMask {
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
		if err := s.includeEnum(enum.FQEN()); err != nil {
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
		parameter.Object.Description = s.renderComment(comments)
	}

	return parameter, nil
}

func (s *Session) addDefaultSuccessResponse(binding *descriptor.Binding, operation *openapiv3.OperationCore) error {
	if operation.Responses == nil {
		operation.Responses = make(map[string]*openapiv3.Ref[openapiv3.Response])
	}

	var schema *openapiv3.Schema
	if binding.ResponseBody != nil {
		// render field schema
		fieldSchema, err := s.addFieldSchema(binding.ResponseBody.FieldPath.Target())
		if err != nil {
			return err
		}
		schema = fieldSchema
	} else {
		// render schema reference
		name, err := s.schemaNameForFQN(binding.Method.ResponseType.FQMN())
		if err != nil {
			return fmt.Errorf(
				"could not find schema name for %q: %w", binding.Method.ResponseType.FQMN(), err)
		}
		schema = s.createSchemaRef(name)
		if err := s.includeMessage(binding.Method.ResponseType.FQMN()); err != nil {
			return err
		}
	}

	mediaType := &openapiv3.MediaType{
		Object: openapiv3.MediaTypeCore{
			Schema:  schema,
			Example: nil,
		},
	}

	response := &openapiv3.Ref[openapiv3.Response]{
		Data: openapiv3.Response{
			Object: openapiv3.ResponseCore{
				Description: defaultSuccessfulResponse,
				Content:     map[string]*openapiv3.MediaType{},
			},
		},
	}

	if binding.Method.GetServerStreaming() {
		response.Data.Object.Description += streamingResponsesDescription
		if binding.StreamConfig.AllowSSE {
			response.Data.Object.Content[mimeTypeSSE] = mediaType
		}
		if binding.StreamConfig.AllowChunkedTransfer {
			response.Data.Object.Content[mimeTypeJSON] = mediaType
		}
	} else {
		response.Data.Object.Content[mimeTypeJSON] = mediaType
	}

	operation.Responses[httpStatusOK] = response
	return nil
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

func (s *Session) updatePathParameterAliasesMap(table map[string]string, binding *descriptor.Binding) map[string]string {
	for _, param := range binding.PathParameters {
		paramPath := param.String()
		if table != nil && table[paramPath] == "" {
			continue
		}

		// NOTE: It is expensive to use the existing function to build the full field customization object.
		// thus a cheaper logic is used here to look up path parameter name.
		field := param.FieldPath.Target()
		if msg := s.messages[field.Message.FQMN()]; msg != nil && msg.Fields != nil {
			if paramName := msg.Fields[field.GetName()].GetConfig().GetPathParamName(); paramName != "" {
				table = addPathParameterAlias(table, paramPath, paramName)
				continue
			}
		}

		fieldConfig, ok := proto.GetExtension(field.Options, api.E_OpenapiField).(*openapi.Schema)
		if !ok || fieldConfig == nil {
			continue
		}

		if paramName := fieldConfig.GetConfig().GetPathParamName(); paramName != "" {
			table = addPathParameterAlias(table, paramPath, paramName)
		}
	}
	return table
}

func addPathParameterAlias(table map[string]string, key, value string) map[string]string {
	if table == nil {
		table = map[string]string{
			key: value,
		}
	} else {
		table[key] = value
	}
	return table
}

func renderPath(binding *descriptor.Binding, aliasMap map[string]string) string {
	writer := &strings.Builder{}

	if len(binding.PathTemplate.Segments) == 0 {
		return "/"
	}

	for _, segment := range binding.PathTemplate.Segments {
		switch segment.Type {
		case httprule.SegmentTypeLiteral:
			writer.WriteString("/" + segment.Value)
		case httprule.SegmentTypeSelector, httprule.SegmentTypeCatchAllSelector:
			if aliasMap != nil {
				if name, ok := aliasMap[segment.Value]; ok {
					writer.WriteString("/{" + name + "}")
					continue
				}
			}
			writer.WriteString("/{" + segment.Value + "}")
		case httprule.SegmentTypeWildcard:
			_, _ = fmt.Fprintf(writer, "/?")
		default:
			_, _ = fmt.Fprintf(writer, "/<!?:%s>", segment.Value)
		}
	}

	return writer.String()
}

func (g *Generator) getCustomizedMethodOperation(method *descriptor.Method) (*openapiv3.Operation, error) {
	var operation *openapiv3.Operation

	if serviceConfig, ok := g.services[method.Service.FQSN()]; ok && serviceConfig.Methods != nil {
		methodConfig, ok := serviceConfig.Methods[method.GetName()]
		if ok {
			result, err := openapimap.Operation(methodConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to map operation: %w", err)
			}

			operation = result
		}
	}

	protoConfig, ok := proto.GetExtension(method.Options, api.E_OpenapiOperation).(*openapi.Operation)
	if ok && protoConfig != nil {
		result, err := openapimap.Operation(protoConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to map operation from proto config: %w", err)
		}

		if operation == nil {
			operation = result
		} else if err := g.mergeObjects(operation, result); err != nil {
			return nil, err
		}
	}

	return operation, nil
}
