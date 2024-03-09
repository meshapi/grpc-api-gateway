package genopenapi

import (
	"fmt"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"google.golang.org/protobuf/types/known/structpb"
)

func mapDocument(doc *openapi.Document) (*openapiv3.Document, error) {
	if doc == nil {
		return nil, nil
	}

	extensions, err := mapExtensions(doc.Extensions)
	if err != nil {
		return nil, err
	}

	info, err := mapInfo(doc.Info)
	if err != nil {
		return nil, fmt.Errorf("invalid info object: %w", err)
	}

	result := &openapiv3.Document{
		Object: openapiv3.DocumentCore{
			Info: info,
		},
		Extensions: extensions,
	}

	result.Object.ExternalDocumentation, err = mapExternalDoc(doc.ExternalDocs)
	if err != nil {
		return nil, fmt.Errorf("invalid external doc: %w", err)
	}

	result.Object.Servers, err = mapServers(doc.Servers)
	if err != nil {
		return nil, fmt.Errorf("invalid servers list: %w", err)
	}

	if doc.Security != nil {
		result.Object.Security = map[string][]string{}
		for _, security := range doc.Security {
			result.Object.Security[security.Name] = security.Scopes
		}
	}

	result.Object.Tags, err = mapTags(doc.Tags)
	if err != nil {
		return nil, fmt.Errorf("invalid tags list: %w", err)
	}

	result.Object.Components, err = mapComponents(doc.Components)
	if err != nil {
		return nil, fmt.Errorf("invalid components object: %w", err)
	}

	return result, nil
}

func mapExternalDoc(doc *openapi.ExternalDocumentation) (*openapiv3.ExternalDocumentation, error) {
	if doc == nil {
		return nil, nil
	}

	extensions, err := mapExtensions(doc.Extensions)
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

func mapExtensions(table map[string]*structpb.Value) (map[string]any, error) {
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

func mapInfo(info *openapi.Info) (*openapiv3.Info, error) {
	if info == nil {
		return nil, nil
	}

	extensions, err := mapExtensions(info.Extensions)
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
		extensions, err = mapExtensions(info.Contact.Extensions)
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
		extensions, err = mapExtensions(info.License.Extensions)
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

func mapTags(tags []*openapi.Tag) ([]openapiv3.Tag, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	// defining these variables outside of the for loop to reuse them.
	var extensions openapiv3.Extensions
	var externalDocs *openapiv3.ExternalDocumentation
	var err error

	result := make([]openapiv3.Tag, len(tags))
	for index, tag := range tags {
		extensions, err = mapExtensions(tag.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid tag at index %d: %w", index, err)
		}

		externalDocs, err = mapExternalDoc(tag.ExternalDocs)
		if err != nil {
			return nil, fmt.Errorf("invalid external doc in tag at index %d: %w", index, err)
		}

		result[index] = openapiv3.Tag{
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

func mapServers(servers []*openapi.Server) ([]openapiv3.Server, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	// defining these variables outside of the for loop to reuse them.
	var extensions, serverVarExtensions openapiv3.Extensions
	var err error

	result := make([]openapiv3.Server, len(servers))
	for index, server := range servers {
		extensions, err = mapExtensions(server.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid server at index %d: %w", index, err)
		}

		var vars map[string]openapiv3.ServerVariable
		if server.Variables != nil {
			vars = map[string]openapiv3.ServerVariable{}
			for name, serverVariable := range server.Variables {
				serverVarExtensions, err = mapExtensions(serverVariable.Extensions)
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

		result[index] = openapiv3.Server{
			Object: openapiv3.ServerCore{
				URL:         server.Url,
				Description: server.Description,
				Variables:   vars,
			},
			Extensions: extensions,
		}
	}

	return result, nil
}

func mapSchema(schema *openapi.Schema) (*openapiv3.Schema, error) {
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
			Type:             mapType(schema.Types),
			Description:      schema.Description,
			Default:          schema.Default.AsInterface(),
			ReadOnly:         schema.ReadyOnly,
			WriteOnly:        schema.WriteOnly,
			Format:           schema.Format,
		},
	}

	var err error

	result.Object.ExternalDocumentation, err = mapExternalDoc(schema.ExternalDocs)
	if err != nil {
		return nil, fmt.Errorf("invalid external doc object: %w", err)
	}

	result.Object.Items, err = mapSchemaItemSpec(schema.Items)
	if err != nil {
		return nil, fmt.Errorf("invalid schema items list: %w", err)
	}

	result.Object.Properties, err = mapSchemaMap(schema.Properties)
	if err != nil {
		return nil, fmt.Errorf("invalid properties object: %w", err)
	}

	result.Object.AdditionalProperties, err = mapSchema(schema.AdditionalProperties)
	if err != nil {
		return nil, fmt.Errorf("invalid additionalProperties object: %w", err)
	}

	result.Object.AllOf, err = mapSchemaList(schema.AllOf)
	if err != nil {
		return nil, fmt.Errorf("invalid allOf object: %w", err)
	}

	result.Object.AnyOf, err = mapSchemaList(schema.AnyOf)
	if err != nil {
		return nil, fmt.Errorf("invalid anyOf object: %w", err)
	}

	result.Object.OneOf, err = mapSchemaList(schema.OneOf)
	if err != nil {
		return nil, fmt.Errorf("invalid oneOf object: %w", err)
	}

	result.Object.Not, err = mapSchema(schema.Not)
	if err != nil {
		return nil, fmt.Errorf("invalid not object: %w", err)
	}

	result.Object.Discriminator, err = mapDiscrimnator(schema.Discriminator)
	if err != nil {
		return nil, fmt.Errorf("invalid discriminator object: %w", err)
	}

	result.Object.Examples = mapExamples(schema.Examples)

	// Add extra keys, note that in this specific case 'x-' prefix is not necessary.
	if schema.Extra != nil {
		result.Extensions = make(openapiv3.Extensions, len(schema.Extra))
		for key, value := range schema.Extra {
			result.Extensions[key] = value.AsInterface()
		}
	}

	return result, nil
}

func mapDiscrimnator(value *openapi.Discriminator) (*openapiv3.Discriminator, error) {
	if value == nil {
		return nil, nil
	}

	extensions, err := mapExtensions(value.Extensions)
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

func mapType(types []openapi.SchemaDataType) openapiv3.TypeSet {
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

func mapSchemaItemSpec(spec *openapi.Schema_Item) (*openapiv3.ItemSpec, error) {
	if spec == nil {
		return nil, nil
	}

	switch spec := spec.Value.(type) {
	case *openapi.Schema_Item_List:
		if spec.List == nil {
			return nil, nil
		}

		schemas, err := mapSchemaList(spec.List.Items)
		if err != nil {
			return nil, err
		}

		return &openapiv3.ItemSpec{Items: schemas}, nil

	case *openapi.Schema_Item_Schema:
		schema, err := mapSchema(spec.Schema)
		if err != nil {
			return nil, fmt.Errorf("invalid schema: %w", err)
		}
		return &openapiv3.ItemSpec{Schema: schema}, nil
	}

	return nil, nil
}

func mapSchemaList(schemas []*openapi.Schema) ([]*openapiv3.Schema, error) {
	if len(schemas) == 0 {
		return nil, nil
	}

	result := make([]*openapiv3.Schema, len(schemas))
	for index, schemaFromProto := range schemas {
		schema, err := mapSchema(schemaFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid schema at index %d: %w", index, err)
		}
		result[index] = schema
	}

	return result, nil
}

func mapSchemaMap(spec map[string]*openapi.Schema) (map[string]*openapiv3.Schema, error) {
	if spec == nil {
		return nil, nil
	}

	result := map[string]*openapiv3.Schema{}

	for key, schemaFromProto := range spec {
		schema, err := mapSchema(schemaFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid schema object for key %q: %w", key, err)
		}
		result[key] = schema
	}

	return result, nil
}

func mapExamples(items []*structpb.Value) []any {
	if items == nil {
		return nil
	}

	result := make([]any, len(items))
	for index, item := range items {
		result[index] = item.AsInterface()
	}

	return nil
}

func mapComponents(components *openapi.Components) (*openapiv3.Components, error) {
	if components == nil {
		return nil, nil
	}

	var err error

	result := &openapiv3.Components{
		Object: openapiv3.ComponentsCore{},
	}

	result.Object.Schemas, err = mapSchemaMap(components.Schemas)
	if err != nil {
		return nil, fmt.Errorf("invalid schemas object: %w", err)
	}

	result.Extensions, err = mapExtensions(components.Extensions)
	if err != nil {
		return nil, err
	}

	return result, nil
}
