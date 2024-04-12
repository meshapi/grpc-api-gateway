package genopenapi

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/internal"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/pathfilter"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v3"
)

type Session struct {
	Document *openapiv3.Document

	generator       *Generator
	includedSchemas map[string]struct{}
}

func (g *Generator) newSession(doc *openapiv3.Document) *Session {
	return &Session{
		Document:        doc,
		generator:       g,
		includedSchemas: make(map[string]struct{}),
	}
}

func (s *Session) includeMessage(location, fqmn string) error {
	// TODO: we may need to do less object nesting here.

	// TODO: we may need to build the full FQMN here.
	// if schema is already added, skip processing it.
	if _, ok := s.includedSchemas[fqmn]; ok {
		return nil
	}

	schema, err := s.generator.getSchemaForMessage(location, fqmn)
	if err != nil {
		return fmt.Errorf("failed to render proto message %q to OpenAPI schema: %w", fqmn, err)
	}

	if s.Document.Object.Components.Object.Schemas == nil {
		s.Document.Object.Components.Object.Schemas = make(map[string]*openapiv3.Schema)
	}

	name, ok := s.generator.schemaNames[fqmn]
	if !ok {
		return fmt.Errorf("unrecognized message: %s", fqmn)
	}

	s.includedSchemas[fqmn] = struct{}{}
	s.Document.Object.Components.Object.Schemas[name] = schema.Schema

	if err := s.includeDependencies(location, schema.Dependencies); err != nil {
		return err
	}

	return nil
}

func (s *Session) includeEnum(location, fqen string) error {
	// TODO: we may need to do less object nesting here.

	// TODO: we may need to build the full FQMN here.
	// if schema is already added, skip processing it.
	if _, ok := s.includedSchemas[fqen]; ok {
		return nil
	}

	schema, err := s.generator.getSchemaForEnum(location, fqen)
	if err != nil {
		return fmt.Errorf("failed to render proto enum %q to OpenAPI schema: %w", fqen, err)
	}

	if s.Document.Object.Components.Object.Schemas == nil {
		s.Document.Object.Components.Object.Schemas = make(map[string]*openapiv3.Schema)
	}

	name, ok := s.generator.schemaNames[fqen]
	if !ok {
		return fmt.Errorf("unrecognized message: %s", fqen)
	}

	s.includedSchemas[fqen] = struct{}{}
	s.Document.Object.Components.Object.Schemas[name] = schema

	return nil
}

func (s *Session) includeDependencies(location string, dependencies internal.SchemaDependencyStore) error {
	for fqn, dependency := range dependencies {
		switch dependency.Kind {
		case internal.DependencyKindMessage:
			if err := s.includeMessage(location, fqn); err != nil {
				return fmt.Errorf("failed to include message dependency %q: %w", fqn, err)
			}
		case internal.DependencyKindEnum:
			if err := s.includeEnum(location, fqn); err != nil {
				return fmt.Errorf("failed to include enum dependency %q: %w", fqn, err)
			}
		}
	}

	return nil
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

	if err := doc.Object.Validate(); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI file: %w", err)
	}

	content := &bytes.Buffer{}
	var extension string

	switch g.OutputFormat {
	case OutputFormatYAML:
		encoder := yaml.NewEncoder(content)
		encoder.SetIndent(2)
		if err := encoder.Encode(doc); err != nil {
			return nil, fmt.Errorf("failed to marshal OpenAPI to yaml: %w", err)
		}
		extension = extYAML
	case OutputFormatJSON:
		encoder := json.NewEncoder(content)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(doc); err != nil {
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
func (g *Generator) renderMessageSchemaWithFilter(
	message *descriptor.Message, filter *pathfilter.Instance) (internal.OpenAPISchema, error) {

	originalSchema, err := g.getSchemaForMessage("", message.FQMN())
	if err != nil {
		return internal.OpenAPISchema{}, fmt.Errorf("failed to get schema for %q: %w", message.FQMN(), err)
	}

	schemaCopy := *originalSchema.Schema
	schemaCopy.Object.Properties = make(map[string]*openapiv3.Schema)
	result := internal.OpenAPISchema{
		Schema:       &schemaCopy,
		Dependencies: originalSchema.Dependencies.Copy(),
	}

	// TODO: need to update required list as well.
	for _, field := range message.Fields {
		impacted, instance := filter.HasString(field.GetName())

		switch {
		case !impacted:
			// If unimpacted, just use the same value.
			fieldName := g.fieldName(field)
			result.Schema.Object.Properties[fieldName] = originalSchema.Schema.Object.Properties[fieldName]
		case instance.Excluded:
			// path parameter cannot be excluded and be a non-scalar type so there is only one scenario in which dependencies
			// need to get updated, if the type is an enum type.
			if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
				enum, err := g.registry.LookupEnum(message.FQMN(), field.GetTypeName())
				if err != nil {
					return internal.OpenAPISchema{}, fmt.Errorf("failed to find enum %q: %w", field.GetTypeName(), err)
				}
				result.Dependencies.Drop(enum.FQEN())
			}
		default:
			underlyingMessage, err := g.registry.LookupMessage(message.FQMN(), field.GetTypeName())
			if err != nil {
				return internal.OpenAPISchema{}, fmt.Errorf("failed to find message %q: %w", field.GetTypeName(), err)
			}
			modifiedSchema, err := g.renderMessageSchemaWithFilter(underlyingMessage, instance)
			if err != nil {
				return internal.OpenAPISchema{}, fmt.Errorf("failed to render filtered message %q: %w", underlyingMessage.FQMN(), err)
			}
			result.Schema.Object.Properties[g.fieldName(field)] = modifiedSchema.Schema
			result.Dependencies.Drop(underlyingMessage.FQMN())
		}
	}

	return result, nil
}
