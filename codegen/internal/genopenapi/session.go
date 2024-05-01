package genopenapi

import (
	"bytes"
	"encoding/json"
	"fmt"

	"dario.cat/mergo"
	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/internal"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/pathfilter"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v3"
)

type Session struct {
	*Generator
	Document *openapiv3.Document

	includedSchemas                      map[string]struct{}
	includedDefaultErrorStatusDependency bool
	hasAnyGeneratedObject                bool
}

func (g *Generator) newSession(doc *openapiv3.Document) *Session {
	return &Session{
		Document:        doc,
		Generator:       g,
		includedSchemas: make(map[string]struct{}),
	}
}

func (s *Session) includeMessage(fqmn string) error {
	s.hasAnyGeneratedObject = true
	if _, ok := s.includedSchemas[fqmn]; ok {
		return nil
	}

	schema, err := s.getSchemaForMessage("", fqmn)
	if err != nil {
		return fmt.Errorf("failed to render proto message %q to OpenAPI schema: %w", fqmn, err)
	}

	schemas := s.Document.Object.Components.Object.Schemas
	if schemas == nil {
		schemas = make(map[string]*openapiv3.Schema)
		s.Document.Object.Components.Object.Schemas = schemas
	}

	name, ok := s.schemaNames[fqmn]
	if !ok {
		return fmt.Errorf("unrecognized message: %s", fqmn)
	}

	s.includedSchemas[fqmn] = struct{}{}
	schemas[name] = schema.Schema

	if err := s.includeDependencies(schema.Dependencies); err != nil {
		return err
	}

	return nil
}

func (s *Session) includeEnum(fqen string) error {
	s.hasAnyGeneratedObject = true
	if _, ok := s.includedSchemas[fqen]; ok {
		return nil
	}

	schema, err := s.getSchemaForEnum("", fqen)
	if err != nil {
		return fmt.Errorf("failed to render proto enum %q to OpenAPI schema: %w", fqen, err)
	}

	schemas := s.Document.Object.Components.Object.Schemas
	if schemas == nil {
		schemas = make(map[string]*openapiv3.Schema)
		s.Document.Object.Components.Object.Schemas = schemas
	}

	name, ok := s.schemaNames[fqen]
	if !ok {
		return fmt.Errorf("unrecognized message: %s", fqen)
	}

	s.includedSchemas[fqen] = struct{}{}
	schemas[name] = schema

	return nil
}

func (s *Session) includeDependency(location string, dependency internal.SchemaDependency) error {
	switch dependency.Kind {
	case internal.DependencyKindMessage:
		return s.includeMessage(dependency.FQN)
	case internal.DependencyKindEnum:
		return s.includeEnum(dependency.FQN)
	}
	return fmt.Errorf("unexpected dependency kind: %v", dependency.Kind)
}

func (s *Session) includeDependencies(dependencies internal.SchemaDependencyStore) error {
	for fqn, dependency := range dependencies {
		switch dependency.Kind {
		case internal.DependencyKindMessage:
			if err := s.includeMessage(fqn); err != nil {
				return fmt.Errorf("failed to include message dependency %q: %w", fqn, err)
			}
		case internal.DependencyKindEnum:
			if err := s.includeEnum(fqn); err != nil {
				return fmt.Errorf("failed to include enum dependency %q: %w", fqn, err)
			}
		}
	}

	return nil
}

func (s *Session) defaultResponsesForService(service *descriptor.Service) (internal.DefaultResponses, error) {
	var fromConfig, fromProto map[string]*openapi.Response
	if doc := s.services[service.FQSN()]; doc != nil {
		fromConfig = doc.GetDocument().GetConfig().GetDefaultResponses()
	}
	serviceSpec, ok := proto.GetExtension(service.Options, api.E_OpenapiServiceDoc).(*openapi.Document)
	if ok && serviceSpec != nil {
		fromProto = serviceSpec.GetConfig().GetDefaultResponses()
	}
	serviceDefaultResponseSpec := internal.MergeDefaultResponseSpec(fromProto, fromConfig)

	if serviceDefaultResponseSpec == nil {
		file := s.LookupFile(service.File)
		if err := s.resolveRefReferencesInDefaultResponses(file.DefaultResponses); err != nil {
			return nil, err
		}
		return file.DefaultResponses, nil
	}

	responses, err := internal.MapDefaultResponses(serviceDefaultResponseSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to map default responses to OpenAPI responses: %w", err)
	}

	// Review all of the schemas, if any $ref is used, update and replace them with OpenAPI references.
	result := internal.MergeDefaultResponses(responses, s.LookupFile(service.File).DefaultResponses)
	if err := s.resolveRefReferencesInDefaultResponses(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Session) resolveRefReferencesInDefaultResponses(responses internal.DefaultResponses) error {
	if len(responses) == 0 {
		return nil
	}

	// Review all of the schemas, if any $ref is used, update and replace them with OpenAPI references.
	for _, response := range responses {
		if response.Processed {
			continue
		}
		if response.Response.Reference != nil {
			continue
		}

		for _, header := range response.Response.Data.Object.Headers {
			if ref := header.Data.Object.Schema.Object.Ref; ref != "" && ref[0] == '.' {
				record, ok := s.schemaNames[ref]
				if !ok {
					continue
				}

				header.Data.Object.Schema.Object.Ref = refPrefix + record
				var err error
				response.Dependencies, err = s.addSchemaDependencyForFQN(response.Dependencies, ref)
				if err != nil {
					return err
				}
			}
		}

		for _, content := range response.Response.Data.Object.Content {
			schema := content.Object.Schema
			if schema == nil {
				continue
			}

			if ref := schema.Object.Ref; ref != "" && schema.Object.Ref[0] == '.' {
				record, ok := s.schemaNames[ref]
				if !ok {
					continue
				}

				schema.Object.Ref = refPrefix + record
				var err error
				response.Dependencies, err = s.addSchemaDependencyForFQN(response.Dependencies, ref)
				if err != nil {
					return err
				}
			}
		}

		response.Processed = true
	}

	return nil
}

// addSchemaDependencyForFQN produces a schema dependency for an FQN with unknown type.
func (g *Generator) addSchemaDependencyForFQN(
	store internal.SchemaDependencyStore, ref string) (internal.SchemaDependencyStore, error) {
	if _, err := g.registry.LookupMessage("", ref); err == nil {
		return internal.AddDependency(
			store, internal.SchemaDependency{FQN: ref, Kind: internal.DependencyKindMessage}), nil
	} else if _, err := g.registry.LookupEnum("", ref); err == nil {
		return internal.AddDependency(
			store, internal.SchemaDependency{FQN: ref, Kind: internal.DependencyKindEnum}), nil
	} else {
		return nil, fmt.Errorf("could not determine dependency kind for %q", ref)
	}
}

func (g *Generator) mergeDocumentWithSeedFile(doc *openapiv3.Document) (map[string]any, error) {
	seedContent, err := g.readOpenAPISeedFile()
	if err != nil {
		return nil, err
	}
	if err := mergo.Map(&seedContent, doc); err != nil {
		return nil, fmt.Errorf("failed to merge seed content with generated OpenAPI document: %w", err)
	}
	return seedContent, nil
}

func (g *Generator) writeDocument(filePrefix string, doc *openapiv3.Document) (*descriptor.ResponseFile, error) {
	if doc == nil {
		return nil, nil
	}

	doc.Object.OpenAPI = openapiv3.Version
	if doc.Object.Info == nil {
		doc.Object.Info = &openapiv3.Extensible[openapiv3.InfoCore]{
			Object: openapiv3.InfoCore{
				Version: "version not set",
			},
		}
	} else if doc.Object.Info.Object.Version == "" {
		doc.Object.Info.Object.Version = "version not set"
	}

	content := &bytes.Buffer{}
	var extension string

	switch g.OutputFormat {
	case OutputFormatYAML:
		encoder := yaml.NewEncoder(content)
		encoder.SetIndent(2)
		var content any = doc
		if g.OpenAPISeedFile != "" {
			seedContent, err := g.mergeDocumentWithSeedFile(doc)
			if err != nil {
				return nil, err
			}
			content = seedContent
		}
		if err := encoder.Encode(content); err != nil {
			return nil, fmt.Errorf("failed to marshal OpenAPI to yaml: %w", err)
		}
		extension = extYAML
	case OutputFormatJSON:
		encoder := json.NewEncoder(content)
		encoder.SetIndent("", "  ")
		var content any = doc
		if g.OpenAPISeedFile != "" {
			seedContent, err := g.mergeDocumentWithSeedFile(doc)
			if err != nil {
				return nil, err
			}
			content = seedContent
		}
		if err := encoder.Encode(content); err != nil {
			return nil, fmt.Errorf("failed to marshal OpenAPI to json: %w", err)
		}
		extension = extJSON
	default:
		return nil, fmt.Errorf("unexpected output format: %v", g.OutputFormat)
	}

	fileName := filePrefix + "." + extension
	data := content.String()
	return &descriptor.ResponseFile{
		CodeGeneratorResponse_File: &pluginpb.CodeGeneratorResponse_File{
			Name:    &fileName,
			Content: &data,
		},
	}, nil
}

// renderMessageSchemaWithFilter is similar to renderMessageSchema but it removes fields from the pathfilter and
// renders schemas with modified fields. Fields that do not match remain unaffected.
func (s *Session) renderMessageSchemaWithFilter(
	message *descriptor.Message, filter *pathfilter.Instance) (internal.OpenAPISchema, error) {

	originalSchema, err := s.getSchemaForMessage("", message.FQMN())
	if err != nil {
		return internal.OpenAPISchema{}, fmt.Errorf("failed to get schema for %q: %w", message.FQMN(), err)
	}

	schemaCopy := *originalSchema.Schema
	schemaCopy.Object.Properties = make(map[string]*openapiv3.Schema)
	result := internal.OpenAPISchema{
		Schema:       &schemaCopy,
		Dependencies: originalSchema.Dependencies.Copy(),
	}

	hasRequiredSlice := len(schemaCopy.Object.Required) > 0
	var fieldsToRemove map[string]struct{}

	for _, field := range message.Fields {
		impacted, instance := filter.HasString(field.GetName())

		switch {
		case !impacted:
			// If unimpacted, just use the same value.
			fieldName := s.fieldName(field)
			result.Schema.Object.Properties[fieldName] = originalSchema.Schema.Object.Properties[fieldName]
		case instance.Excluded:
			// path parameter cannot be excluded and be a non-scalar type so there is only one scenario in which dependencies
			// need to get updated, if the type is an enum type.
			if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
				enum, err := s.registry.LookupEnum(message.FQMN(), field.GetTypeName())
				if err != nil {
					return internal.OpenAPISchema{}, fmt.Errorf("failed to find enum %q: %w", field.GetTypeName(), err)
				}
				result.Dependencies.Drop(enum.FQEN())
			}
			if hasRequiredSlice {
				fieldsToRemove = s.appendFieldToRemovedFields(fieldsToRemove, field)
			}
		default:
			underlyingMessage, err := s.registry.LookupMessage(message.FQMN(), field.GetTypeName())
			if err != nil {
				return internal.OpenAPISchema{}, fmt.Errorf("failed to find message %q: %w", field.GetTypeName(), err)
			}
			modifiedSchema, err := s.renderMessageSchemaWithFilter(underlyingMessage, instance)
			if err != nil {
				return internal.OpenAPISchema{}, fmt.Errorf("failed to render filtered message %q: %w", underlyingMessage.FQMN(), err)
			}
			if len(modifiedSchema.Schema.Object.Properties) != 0 {
				result.Schema.Object.Properties[s.fieldName(field)] = modifiedSchema.Schema
				result.Dependencies.Drop(underlyingMessage.FQMN())
			} else if hasRequiredSlice {
				// NOTE: Messages that are empty are considered to be deliberately empty so they do not get removed.
				// However, the schemas that become empty as part of this filtering down get removed.
				fieldsToRemove = s.appendFieldToRemovedFields(fieldsToRemove, field)
			}
		}
	}

	if fieldsToRemove != nil {
		schemaCopy.Object.Required = internal.FilteredStringSlice(schemaCopy.Object.Required, fieldsToRemove)
	}

	return result, nil
}

func (s *Session) appendFieldToRemovedFields(table map[string]struct{}, field *descriptor.Field) map[string]struct{} {
	if table == nil {
		return map[string]struct{}{s.fieldName(field): {}}
	}

	table[s.fieldName(field)] = struct{}{}
	return table
}
