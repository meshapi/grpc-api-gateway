// Package openapimap provides code to map openapi config from the proto messages to openapiv3 types.
package openapimap

import (
	"fmt"
	"strings"

	"github.com/meshapi/grpc-api-gateway/api/openapi"
	"github.com/meshapi/grpc-api-gateway/codegen/internal/openapiv3"
	"google.golang.org/protobuf/types/known/structpb"
)

func Document(doc *openapi.Document) (*openapiv3.Document, error) {
	if doc == nil {
		return nil, nil
	}

	extensions, err := Extensions(doc.Extensions)
	if err != nil {
		return nil, err
	}

	info, err := Info(doc.Info)
	if err != nil {
		return nil, fmt.Errorf("invalid info object: %w", err)
	}

	result := &openapiv3.Document{
		Object: openapiv3.DocumentCore{
			Info: info,
		},
		Extensions: extensions,
	}

	result.Object.ExternalDocumentation, err = ExternalDoc(doc.ExternalDocs)
	if err != nil {
		return nil, fmt.Errorf("invalid external doc: %w", err)
	}

	result.Object.Servers, err = ServerSlice(doc.Servers)
	if err != nil {
		return nil, fmt.Errorf("invalid servers list: %w", err)
	}

	result.Object.Security = SecurityRequirementSlice(doc.Security)

	result.Object.Tags, err = Tags(doc.Tags)
	if err != nil {
		return nil, fmt.Errorf("invalid tags list: %w", err)
	}

	result.Object.Components, err = Components(doc.Components)
	if err != nil {
		return nil, fmt.Errorf("invalid components object: %w", err)
	}

	return result, nil
}

func ExternalDoc(doc *openapi.ExternalDocumentation) (*openapiv3.ExternalDocumentation, error) {
	if doc == nil {
		return nil, nil
	}

	extensions, err := Extensions(doc.Extensions)
	if err != nil {
		return nil, err
	}

	return &openapiv3.ExternalDocumentation{
		Object: openapiv3.ExternalDocumentationCore{
			Description: doc.Description,
			URL:         doc.Url,
		},
		Extensions: extensions,
	}, nil
}

func Extensions(table map[string]*structpb.Value) (map[string]any, error) {
	if table == nil {
		return nil, nil
	}

	result := make(map[string]any, len(table))
	for key, val := range table {
		if !strings.HasPrefix(key, "x-") {
			return nil, fmt.Errorf("extension keys must begin with prefix 'x-', value %q is not accepted", key)
		}
		result[key] = val.AsInterface()
	}

	return result, nil
}

func Info(info *openapi.Info) (*openapiv3.Info, error) {
	if info == nil {
		return nil, nil
	}

	extensions, err := Extensions(info.Extensions)
	if err != nil {
		return nil, err
	}

	result := &openapiv3.Info{
		Object: openapiv3.InfoCore{
			Title:          info.Title,
			Summary:        info.Summary,
			Description:    info.Description,
			TermsOfService: info.TermsOfService,
			Version:        info.Version,
		},
		Extensions: extensions,
	}

	if info.Contact != nil {
		extensions, err = Extensions(info.Contact.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid contact object: %w", err)
		}

		result.Object.Contact = &openapiv3.Contact{
			Object: openapiv3.ContactCore{
				Name:  info.Contact.Name,
				URL:   info.Contact.Url,
				Email: info.Contact.Email,
			},
			Extensions: extensions,
		}
	}

	if info.License != nil {
		extensions, err = Extensions(info.License.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid contact object: %w", err)
		}

		result.Object.License = &openapiv3.License{
			Object: openapiv3.LicenseCore{
				Name:       info.License.Name,
				Identifier: info.License.Identifier,
				URL:        info.License.Url,
			},
			Extensions: extensions,
		}
	}

	return result, nil
}

func Tags(tags []*openapi.Tag) ([]*openapiv3.Tag, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	// defining these variables outside of the for loop to reuse them.
	var extensions openapiv3.Extensions
	var externalDocs *openapiv3.ExternalDocumentation
	var err error

	result := make([]*openapiv3.Tag, len(tags))
	for index, tag := range tags {
		extensions, err = Extensions(tag.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid tag at index %d: %w", index, err)
		}

		externalDocs, err = ExternalDoc(tag.ExternalDocs)
		if err != nil {
			return nil, fmt.Errorf("invalid external doc in tag at index %d: %w", index, err)
		}

		result[index] = &openapiv3.Tag{
			Object: openapiv3.TagCore{
				Name:         tag.Name,
				Description:  tag.Description,
				ExternalDocs: externalDocs,
			},
			Extensions: extensions,
		}
	}

	return result, nil
}

func Server(server *openapi.Server) (*openapiv3.Server, error) {
	if server == nil {
		return nil, nil
	}

	extensions, err := Extensions(server.Extensions)
	if err != nil {
		return nil, err
	}

	var vars map[string]openapiv3.ServerVariable
	if server.Variables != nil {
		vars = map[string]openapiv3.ServerVariable{}
		for name, serverVariable := range server.Variables {
			serverVarExtensions, err := Extensions(serverVariable.Extensions)
			if err != nil {
				return nil, fmt.Errorf("invalid server variable %q: %w", name, err)
			}

			vars[name] = openapiv3.ServerVariable{
				Object: openapiv3.ServerVariableCore{
					Enum:        serverVariable.EnumValues,
					Default:     serverVariable.DefaultValue,
					Description: serverVariable.Description,
				},
				Extensions: serverVarExtensions,
			}
		}
	}

	return &openapiv3.Server{
		Object: openapiv3.ServerCore{
			URL:         server.Url,
			Description: server.Description,
			Variables:   vars,
		},
		Extensions: extensions,
	}, nil
}

func ServerSlice(servers []*openapi.Server) ([]*openapiv3.Server, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	result := make([]*openapiv3.Server, len(servers))
	for index, serverFromProto := range servers {
		server, err := Server(serverFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid server object at index %d: %w", index, err)
		}
		result[index] = server
	}

	return result, nil
}

func Schema(schema *openapi.Schema) (*openapiv3.Schema, error) {
	if schema == nil {
		return nil, nil
	}

	result := &openapiv3.Schema{
		Object: openapiv3.SchemaCore{
			Ref:              schema.Ref,
			Schema:           schema.Schema,
			Title:            schema.Title,
			Pattern:          schema.Pattern,
			Required:         schema.Required,
			Enum:             schema.Enum,
			MultipleOf:       schema.MultipleOf,
			Maximum:          schema.Maximum,
			ExclusiveMaximum: schema.ExclusiveMaximum,
			Minimum:          schema.Minimum,
			ExclusiveMinimum: schema.ExclusiveMinimum,
			MaxLength:        schema.MaxLength,
			MinLength:        schema.MinLength,
			MaxItems:         schema.MaxItems,
			MinItems:         schema.MinItems,
			UniqueItems:      schema.UniqueItems,
			MaxProperties:    schema.MaxProperties,
			MinProperties:    schema.MinProperties,
			Type:             Type(schema.Types),
			Description:      schema.Description,
			Default:          schema.Default.AsInterface(),
			ReadOnly:         schema.ReadyOnly,
			WriteOnly:        schema.WriteOnly,
			Format:           schema.Format,
			Deprecated:       schema.Deprecated,
		},
	}

	var err error

	result.Object.ExternalDocumentation, err = ExternalDoc(schema.ExternalDocs)
	if err != nil {
		return nil, fmt.Errorf("invalid external doc object: %w", err)
	}

	result.Object.Items, err = SchemaItemSpec(schema.Items)
	if err != nil {
		return nil, fmt.Errorf("invalid schema items list: %w", err)
	}

	result.Object.Properties, err = SchemaMap(schema.Properties)
	if err != nil {
		return nil, fmt.Errorf("invalid properties object: %w", err)
	}

	result.Object.AdditionalProperties, err = Schema(schema.AdditionalProperties)
	if err != nil {
		return nil, fmt.Errorf("invalid additionalProperties object: %w", err)
	}

	result.Object.AllOf, err = SchemaList(schema.AllOf)
	if err != nil {
		return nil, fmt.Errorf("invalid allOf object: %w", err)
	}

	result.Object.AnyOf, err = SchemaList(schema.AnyOf)
	if err != nil {
		return nil, fmt.Errorf("invalid anyOf object: %w", err)
	}

	result.Object.OneOf, err = SchemaList(schema.OneOf)
	if err != nil {
		return nil, fmt.Errorf("invalid oneOf object: %w", err)
	}

	result.Object.Not, err = Schema(schema.Not)
	if err != nil {
		return nil, fmt.Errorf("invalid not object: %w", err)
	}

	result.Object.Discriminator, err = Discriminator(schema.Discriminator)
	if err != nil {
		return nil, fmt.Errorf("invalid discriminator object: %w", err)
	}

	result.Object.Examples = AnySlice(schema.Examples)

	// Add extra keys, note that in this specific case 'x-' prefix is not necessary.
	if schema.Extra != nil {
		result.Extensions = make(openapiv3.Extensions, len(schema.Extra))
		for key, value := range schema.Extra {
			result.Extensions[key] = value.AsInterface()
		}
	}

	return result, nil
}

func Discriminator(value *openapi.Discriminator) (*openapiv3.Discriminator, error) {
	if value == nil {
		return nil, nil
	}

	extensions, err := Extensions(value.Extensions)
	if err != nil {
		return nil, err
	}

	return &openapiv3.Discriminator{
		Object: openapiv3.DiscriminatorCore{
			PropertyName: value.PropertyName,
			Mapping:      value.Mapping,
		},
		Extensions: extensions,
	}, nil
}

func Type(types []openapi.SchemaDataType) openapiv3.TypeSet {
	if len(types) == 0 {
		return nil
	}

	result := make(openapiv3.TypeSet, len(types))
	for index, typeEnum := range types {
		switch typeEnum {
		case openapi.SchemaDataType_ARRAY:
			result[index] = "array"
		case openapi.SchemaDataType_NULL:
			result[index] = "null"
		case openapi.SchemaDataType_OBJECT:
			result[index] = "object"
		case openapi.SchemaDataType_STRING:
			result[index] = "string"
		case openapi.SchemaDataType_BOOLEAN:
			result[index] = "boolean"
		case openapi.SchemaDataType_NUMBER:
			result[index] = "number"
		}
	}

	return result
}

func SchemaItemSpec(spec *openapi.Schema_Item) (*openapiv3.ItemSpec, error) {
	if spec == nil {
		return nil, nil
	}

	switch spec := spec.Value.(type) {
	case *openapi.Schema_Item_List:
		if spec.List == nil {
			return nil, nil
		}

		schemas, err := SchemaList(spec.List.Items)
		if err != nil {
			return nil, err
		}

		return &openapiv3.ItemSpec{Items: schemas}, nil

	case *openapi.Schema_Item_Schema:
		schema, err := Schema(spec.Schema)
		if err != nil {
			return nil, fmt.Errorf("invalid schema: %w", err)
		}
		return &openapiv3.ItemSpec{Schema: schema}, nil
	}

	return nil, nil
}

func SchemaList(schemas []*openapi.Schema) ([]*openapiv3.Schema, error) {
	if len(schemas) == 0 {
		return nil, nil
	}

	result := make([]*openapiv3.Schema, len(schemas))
	for index, schemaFromProto := range schemas {
		schema, err := Schema(schemaFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid schema at index %d: %w", index, err)
		}
		result[index] = schema
	}

	return result, nil
}

func SchemaMap(spec map[string]*openapi.Schema) (map[string]*openapiv3.Schema, error) {
	if spec == nil {
		return nil, nil
	}

	result := map[string]*openapiv3.Schema{}

	for key, schemaFromProto := range spec {
		schema, err := Schema(schemaFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid schema object for key %q: %w", key, err)
		}
		result[key] = schema
	}

	return result, nil
}

func AnyMap(items map[string]*structpb.Value) map[string]any {
	if items == nil {
		return nil
	}

	result := make(map[string]any, len(items))
	for key, value := range items {
		result[key] = value.AsInterface()
	}

	return result
}

func AnySlice(items []*structpb.Value) []any {
	if items == nil {
		return nil
	}

	result := make([]any, len(items))
	for index, item := range items {
		result[index] = item.AsInterface()
	}

	return result
}

func MakeReference[T any](ref *openapi.Reference) *openapiv3.Ref[T] {
	return &openapiv3.Ref[T]{
		Reference: &openapiv3.Reference{
			Ref:         ref.Uri,
			Summary:     ref.Summary,
			Description: ref.Description,
		},
	}
}

func EncodingMap(encodings map[string]*openapi.Encoding) (map[string]*openapiv3.Encoding, error) {
	if encodings == nil {
		return nil, nil
	}

	result := make(map[string]*openapiv3.Encoding, len(encodings))
	for key, encodingFromProto := range encodings {
		extensions, err := Extensions(encodingFromProto.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid encoding for %q: %w", key, err)
		}

		encoding := &openapiv3.Encoding{
			Object: openapiv3.EncodingCore{
				ContentType:   encodingFromProto.ContentType,
				Style:         encodingFromProto.Style,
				Explode:       encodingFromProto.Explode,
				AllowReserved: encodingFromProto.AllowReserved,
			},
			Extensions: extensions,
		}

		encoding.Object.Headers, err = HeaderMap(encodingFromProto.Headers)
		if err != nil {
			return nil, fmt.Errorf("invalid headers object for %q: %w", key, err)
		}

		result[key] = encoding
	}

	return result, nil
}

func MediaTypes(mediaTypes map[string]*openapi.MediaType) (map[string]*openapiv3.MediaType, error) {
	if mediaTypes == nil {
		return nil, nil
	}

	result := make(map[string]*openapiv3.MediaType, len(mediaTypes))
	for key, mediaTypeFromProto := range mediaTypes {
		extensions, err := Extensions(mediaTypeFromProto.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid media type object for %q: %w", key, err)
		}

		mediaType := &openapiv3.MediaType{
			Object: openapiv3.MediaTypeCore{
				Example: mediaTypeFromProto.Example.AsInterface(),
			},
			Extensions: extensions,
		}

		mediaType.Object.Schema, err = Schema(mediaTypeFromProto.Schema)
		if err != nil {
			return nil, fmt.Errorf("invalid schema object for %q: %w", key, err)
		}

		mediaType.Object.Examples, err = StructuredExampleMap(mediaTypeFromProto.Examples)
		if err != nil {
			return nil, fmt.Errorf("invalid examples list for %q: %w", key, err)
		}

		mediaType.Object.Encoding, err = EncodingMap(mediaTypeFromProto.Encoding)
		if err != nil {
			return nil, fmt.Errorf("invalid encoding object for %q: %w", key, err)
		}

		result[key] = mediaType
	}

	return result, nil
}

func HeaderMap(headerMap map[string]*openapi.Header) (map[string]*openapiv3.Ref[openapiv3.Header], error) {
	if headerMap == nil {
		return nil, nil
	}

	result := make(map[string]*openapiv3.Ref[openapiv3.Header], len(headerMap))
	for key, protoHeader := range headerMap {
		if protoHeader.Ref != nil {
			result[key] = MakeReference[openapiv3.Header](protoHeader.Ref)
			continue
		}

		extensions, err := Extensions(protoHeader.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid header for %q: %w", key, err)
		}

		header := &openapiv3.Ref[openapiv3.Header]{
			Data: openapiv3.Header{
				Object: openapiv3.HeaderCore{
					Description:     protoHeader.Description,
					Required:        protoHeader.Required,
					Deprecated:      protoHeader.Deprecated,
					AllowEmptyValue: protoHeader.AllowEmptyValue,
					Style:           protoHeader.Style,
					Explode:         protoHeader.Explode,
					Example:         protoHeader.Example.AsInterface(),
				},
				Extensions: extensions,
			},
		}

		header.Data.Object.Schema, err = Schema(protoHeader.Schema)
		if err != nil {
			return nil, fmt.Errorf("invalid schema for %q: %w", key, err)
		}

		header.Data.Object.Examples, err = StructuredExampleMap(protoHeader.Examples)
		if err != nil {
			return nil, fmt.Errorf("invalid examples list: %w", err)
		}

		header.Data.Object.Content, err = MediaTypes(protoHeader.Content)
		if err != nil {
			return nil, fmt.Errorf("invalid media types object: %w", err)
		}

		result[key] = header
	}

	return result, nil
}

func Parameter(paramFromProto *openapi.Parameter) (*openapiv3.Ref[openapiv3.Parameter], error) {
	if paramFromProto == nil {
		return nil, nil
	}

	if paramFromProto.Ref != nil {
		return MakeReference[openapiv3.Parameter](paramFromProto.Ref), nil
	}

	extensions, err := Extensions(paramFromProto.Extensions)
	if err != nil {
		return nil, err
	}

	result := &openapiv3.Ref[openapiv3.Parameter]{
		Data: openapiv3.Parameter{
			Object: openapiv3.ParameterCore{
				Name:            paramFromProto.Name,
				In:              paramFromProto.In,
				Description:     paramFromProto.Description,
				Required:        paramFromProto.Required,
				Deprecated:      paramFromProto.Deprecated,
				AllowEmptyValue: paramFromProto.AllowEmptyValue,
				AllowReserved:   paramFromProto.AllowReserved,
				Style:           paramFromProto.Style,
				Explode:         paramFromProto.Explode,
				Example:         paramFromProto.Example.AsInterface(),
			},
			Extensions: extensions,
		},
	}

	result.Data.Object.Schema, err = Schema(paramFromProto.Schema)
	if err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	result.Data.Object.Examples, err = StructuredExampleMap(paramFromProto.Examples)
	if err != nil {
		return nil, fmt.Errorf("invalid examples list: %w", err)
	}

	result.Data.Object.Content, err = MediaTypes(paramFromProto.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid media types object: %w", err)
	}

	return result, nil
}

func ParameterMap(parameterMap map[string]*openapi.Parameter) (map[string]*openapiv3.Ref[openapiv3.Parameter], error) {
	if parameterMap == nil {
		return nil, nil
	}

	result := make(map[string]*openapiv3.Ref[openapiv3.Parameter], len(parameterMap))
	for key, paramFromProto := range parameterMap {
		header, err := Parameter(paramFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter for %q: %w", key, err)
		}

		result[key] = header
	}

	return result, nil
}

func ParameterSlice(parameters []*openapi.Parameter) ([]*openapiv3.Ref[openapiv3.Parameter], error) {
	if len(parameters) == 0 {
		return nil, nil
	}

	result := make([]*openapiv3.Ref[openapiv3.Parameter], len(parameters))
	for index, paramFromProto := range parameters {
		param, err := Parameter(paramFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter at index %d: %w", index, err)
		}
		result[index] = param
	}

	return result, nil
}

func StructuredExampleMap(examples map[string]*openapi.Example) (map[string]*openapiv3.Ref[openapiv3.Example], error) {
	if examples == nil {
		return nil, nil
	}

	result := make(map[string]*openapiv3.Ref[openapiv3.Example], len(examples))
	for key, exampleFromProto := range examples {
		example, err := StructuredExample(exampleFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid example at %q: %w", key, err)
		}

		result[key] = example
	}

	return result, nil
}

func StructuredExample(example *openapi.Example) (*openapiv3.Ref[openapiv3.Example], error) {
	if example == nil {
		return nil, nil
	}

	if example.Ref != nil {
		return MakeReference[openapiv3.Example](example.Ref), nil
	}

	extensions, err := Extensions(example.Extensions)
	if err != nil {
		return nil, err
	}

	result := &openapiv3.Ref[openapiv3.Example]{
		Data: openapiv3.Example{
			Object: openapiv3.ExampleCore{
				Summary:       example.Summary,
				Description:   example.Description,
				Value:         example.Value.AsInterface(),
				ExternalValue: example.ExternalValue,
			},
			Extensions: extensions,
		},
	}

	return result, nil
}

func Response(response *openapi.Response) (*openapiv3.Ref[openapiv3.Response], error) {
	if response == nil {
		return nil, nil
	}

	if response.Ref != nil {
		return MakeReference[openapiv3.Response](response.Ref), nil
	}

	extensions, err := Extensions(response.Extensions)
	if err != nil {
		return nil, err
	}

	result := &openapiv3.Ref[openapiv3.Response]{
		Data: openapiv3.Response{
			Object: openapiv3.ResponseCore{
				Description: response.Description,
			},
			Extensions: extensions,
		},
	}

	result.Data.Object.Headers, err = HeaderMap(response.Headers)
	if err != nil {
		return nil, fmt.Errorf("invalid headers object: %w", err)
	}

	result.Data.Object.Content, err = MediaTypes(response.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid content object: %w", err)
	}

	result.Data.Object.Links, err = LinksMap(response.Links)
	if err != nil {
		return nil, fmt.Errorf("invalid links object: %w", err)
	}

	return result, nil
}

func ResponseMap(responses map[string]*openapi.Response) (map[string]*openapiv3.Ref[openapiv3.Response], error) {
	if responses == nil {
		return nil, nil
	}

	result := make(map[string]*openapiv3.Ref[openapiv3.Response], len(responses))
	for key, responseFromProto := range responses {
		response, err := Response(responseFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid response for %q: %w", key, err)
		}
		result[key] = response
	}

	return result, nil
}

func LinksMap(links map[string]*openapi.Link) (map[string]*openapiv3.Ref[openapiv3.Link], error) {
	if links == nil {
		return nil, nil
	}

	result := make(map[string]*openapiv3.Ref[openapiv3.Link], len(links))
	for key, linkFromProto := range links {
		if linkFromProto.Ref != nil {
			result[key] = MakeReference[openapiv3.Link](linkFromProto.Ref)
			continue
		}

		extensions, err := Extensions(linkFromProto.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid link for %q: %w", key, err)
		}

		link := &openapiv3.Ref[openapiv3.Link]{
			Data: openapiv3.Link{
				Object: openapiv3.LinkCore{
					Parameters:  AnyMap(linkFromProto.Parameters),
					RequestBody: linkFromProto.RequestBody.AsInterface(),
					Description: linkFromProto.Description,
				},
				Extensions: extensions,
			},
		}

		switch operation := linkFromProto.Operation.(type) {
		case *openapi.Link_OperationId:
			link.Data.Object.OperationID = operation.OperationId
		case *openapi.Link_OperationRef:
			link.Data.Object.OperationRef = operation.OperationRef
		}

		link.Data.Object.Server, err = Server(linkFromProto.Server)
		if err != nil {
			return nil, fmt.Errorf("invalid server object for %q: %w", key, err)
		}

		result[key] = link
	}

	return result, nil
}

func RequestBody(requestBody *openapi.RequestBody) (*openapiv3.Ref[openapiv3.RequestBody], error) {
	if requestBody == nil {
		return nil, nil
	}

	if requestBody.Ref != nil {
		return MakeReference[openapiv3.RequestBody](requestBody.Ref), nil
	}

	extensions, err := Extensions(requestBody.Extensions)
	if err != nil {
		return nil, err
	}

	result := &openapiv3.Ref[openapiv3.RequestBody]{
		Data: openapiv3.RequestBody{
			Object: openapiv3.RequestBodyCore{
				Description: requestBody.Description,
				Required:    requestBody.Required,
			},
			Extensions: extensions,
		},
	}

	result.Data.Object.Content, err = MediaTypes(requestBody.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid content object: %w", err)
	}

	return result, nil
}

func RequestBodyMap(
	requestBodies map[string]*openapi.RequestBody) (map[string]*openapiv3.Ref[openapiv3.RequestBody], error) {

	if requestBodies == nil {
		return nil, nil
	}

	result := make(map[string]*openapiv3.Ref[openapiv3.RequestBody], len(requestBodies))
	for key, requestBodyFromProto := range requestBodies {
		requestBody, err := RequestBody(requestBodyFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid request body for %q: %w", key, err)
		}

		result[key] = requestBody
	}

	return result, nil
}

func OAuthFlow(flow *openapi.SecurityScheme_OAuthFlow) (*openapiv3.OAuthFlow, error) {
	if flow == nil {
		return nil, nil
	}

	extensions, err := Extensions(flow.Extensions)
	if err != nil {
		return nil, err
	}

	return &openapiv3.OAuthFlow{
		Object: openapiv3.OAuthFlowCore{
			AuthorizationURL: flow.AuthorizationUrl,
			TokenURL:         flow.TokenUrl,
			RefreshURL:       flow.RefreshUrl,
			Scopes:           flow.Scopes,
		},
		Extensions: extensions,
	}, nil
}

func OAuthFlows(flows *openapi.SecurityScheme_OAuthFlows) (*openapiv3.OAuthFlows, error) {
	if flows == nil {
		return nil, nil
	}

	extensions, err := Extensions(flows.Extensions)
	if err != nil {
		return nil, err
	}

	result := &openapiv3.OAuthFlows{
		Object:     openapiv3.OAuthFlowsCore{},
		Extensions: extensions,
	}

	result.Object.Implicit, err = OAuthFlow(flows.Implicit)
	if err != nil {
		return nil, fmt.Errorf("invalid implicit oauth flow: %w", err)
	}

	result.Object.Password, err = OAuthFlow(flows.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid password oauth flow: %w", err)
	}

	result.Object.AuthorizationCode, err = OAuthFlow(flows.AuthorizationCode)
	if err != nil {
		return nil, fmt.Errorf("invalid authorizationCode oauth flow: %w", err)
	}

	result.Object.ClientCredentials, err = OAuthFlow(flows.ClientCredentials)
	if err != nil {
		return nil, fmt.Errorf("invalid clientCredentials oauth flow: %w", err)
	}

	return result, nil
}

func SecurityScheme(scheme *openapi.SecurityScheme) (*openapiv3.Ref[openapiv3.SecurityScheme], error) {
	if scheme == nil {
		return nil, nil
	}

	if scheme.Ref != nil {
		return MakeReference[openapiv3.SecurityScheme](scheme.Ref), nil
	}

	extensions, err := Extensions(scheme.Extensions)
	if err != nil {
		return nil, err
	}

	result := &openapiv3.Ref[openapiv3.SecurityScheme]{
		Data: openapiv3.SecurityScheme{
			Object: openapiv3.SecuritySchemeCore{
				Type:             scheme.Type,
				Description:      scheme.Description,
				Name:             scheme.Name,
				In:               scheme.In,
				Scheme:           scheme.Scheme,
				BearerFormat:     scheme.BearerFormat,
				OpenIDConnectURL: scheme.OpenIdConnectUrl,
			},
			Extensions: extensions,
		},
	}

	result.Data.Object.Flows, err = OAuthFlows(scheme.Flows)
	if err != nil {
		return nil, fmt.Errorf("invalid flows object: %w", err)
	}

	return result, nil
}

func SecuritySchemeMap(
	securitySchemes map[string]*openapi.SecurityScheme) (map[string]*openapiv3.Ref[openapiv3.SecurityScheme], error) {
	if securitySchemes == nil {
		return nil, nil
	}

	result := make(map[string]*openapiv3.Ref[openapiv3.SecurityScheme], len(securitySchemes))
	for key, securitySchemeFromProto := range securitySchemes {
		securityScheme, err := SecurityScheme(securitySchemeFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid securityScheme for %q: %w", key, err)
		}

		result[key] = securityScheme
	}

	return result, nil
}

func Components(components *openapi.Components) (*openapiv3.Components, error) {
	if components == nil {
		return nil, nil
	}

	var err error

	result := &openapiv3.Components{
		Object: openapiv3.ComponentsCore{},
	}

	result.Object.Schemas, err = SchemaMap(components.Schemas)
	if err != nil {
		return nil, fmt.Errorf("invalid schemas object: %w", err)
	}

	result.Object.Responses, err = ResponseMap(components.Responses)
	if err != nil {
		return nil, fmt.Errorf("invalid responses object: %w", err)
	}

	result.Object.Parameters, err = ParameterMap(components.Parameters)
	if err != nil {
		return nil, fmt.Errorf("invalid paramters object: %w", err)
	}

	result.Object.Examples, err = StructuredExampleMap(components.Examples)
	if err != nil {
		return nil, fmt.Errorf("invalid examples object: %w", err)
	}

	result.Object.RequestBodies, err = RequestBodyMap(components.RequestBodies)
	if err != nil {
		return nil, fmt.Errorf("invalid requestBodies object: %w", err)
	}

	result.Object.Headers, err = HeaderMap(components.Headers)
	if err != nil {
		return nil, fmt.Errorf("invalid headers object: %w", err)
	}

	result.Object.SecuritySchemes, err = SecuritySchemeMap(components.SecuritySchemes)
	if err != nil {
		return nil, fmt.Errorf("invalid securitySchemes object: %w", err)
	}

	result.Object.Links, err = LinksMap(components.Links)
	if err != nil {
		return nil, fmt.Errorf("invalid links object: %w", err)
	}

	result.Extensions, err = Extensions(components.Extensions)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func Operation(operation *openapi.Operation) (*openapiv3.Operation, error) {
	if operation == nil {
		return nil, nil
	}

	result := &openapiv3.Operation{
		Object: openapiv3.OperationCore{
			Tags:        operation.Tags,
			Summary:     operation.Summary,
			Description: operation.Description,
			OperationID: operation.OperationId,
			Deprecated:  operation.Deprecated,
			Security:    SecurityRequirementSlice(operation.Security),
		},
	}
	var err error

	result.Extensions, err = Extensions(operation.Extensions)
	if err != nil {
		return nil, err
	}

	result.Object.ExternalDocs, err = ExternalDoc(operation.ExternalDocs)
	if err != nil {
		return nil, fmt.Errorf("invalid external doc: %w", err)
	}

	result.Object.Parameters, err = ParameterSlice(operation.Parameters)
	if err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result.Object.Responses, err = ResponseMap(operation.Responses)
	if err != nil {
		return nil, fmt.Errorf("failed to map responses: %w", err)
	}

	result.Object.Servers, err = ServerSlice(operation.Servers)
	if err != nil {
		return nil, fmt.Errorf("failed to map servers: %w", err)
	}

	return result, nil
}

func SecurityRequirementSlice(items []*openapi.SecurityRequirement) []map[string][]string {
	if len(items) == 0 {
		return nil
	}

	result := make([]map[string][]string, len(items))
	for index, security := range items {
		result[index] = map[string][]string{
			security.Name: security.Scopes,
		}
	}

	return result
}
