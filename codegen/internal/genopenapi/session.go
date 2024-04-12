package genopenapi

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/pathfilter"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
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

func (s *Session) includeDependencies(location string, dependencies []schemaDependency) error {
	for _, dependency := range dependencies {
		switch dependency.Kind {
		case dependencyKindMessage:
			if err := s.includeMessage(location, dependency.FQN); err != nil {
				return fmt.Errorf("failed to include message dependency %q: %w", dependency.FQN, err)
			}
		case dependencyKindEnum:
			if err := s.includeEnum(location, dependency.FQN); err != nil {
				return fmt.Errorf("failed to include enum dependency %q: %w", dependency.FQN, err)
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
		extension = "yaml"
	case OutputFormatJSON:
		encoder := json.NewEncoder(content)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(doc); err != nil {
			return nil, fmt.Errorf("failed to marshal OpenAPI to json: %w", err)
		}
		extension = "json"
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
	message *descriptor.Message, filter *pathfilter.Instance) (openAPISchemaConfig, error) {

	originalSchema, err := g.getSchemaForMessage("", message.FQMN())
	if err != nil {
		return openAPISchemaConfig{}, fmt.Errorf("failed to get schema for %q: %w", message.FQMN(), err)
	}

	schemaCopy := *originalSchema.Schema
	schemaCopy.Object.Properties = make(map[string]*openapiv3.Schema)
	result := openAPISchemaConfig{
		Schema:       &schemaCopy,
		Dependencies: originalSchema.Dependencies,
	}

	// TODO: Update dependencies so they are a map.
	// TODO: need to update required list as well.
	for _, field := range message.Fields {
		impacted, instance := filter.HasString(field.GetName())

		switch {
		case !impacted:
			// If unimpacted, just use the same value.
			switch g.Options.FieldNameMode {
			case FieldNameModeJSON:
				result.Schema.Object.Properties[field.GetJsonName()] = originalSchema.Schema.Object.Properties[field.GetJsonName()]
			case FieldNameModeProto:
				result.Schema.Object.Properties[field.GetName()] = originalSchema.Schema.Object.Properties[field.GetName()]
			}
		case instance.Excluded:
			// if the field needs to get excluded.
			// TODO: remove the dropped required field.
			//if len(result.Schema.Object.Required) > 0 {
			//}
		default:
			underlyingMessage, err := g.registry.LookupMessage(message.FQMN(), field.GetTypeName())
			if err != nil {
				return openAPISchemaConfig{}, fmt.Errorf("failed to find message %q: %w", field.GetTypeName(), err)
			}
			modifiedSchema, err := g.renderMessageSchemaWithFilter(underlyingMessage, instance)
			if err != nil {
				return openAPISchemaConfig{}, fmt.Errorf("failed to render filtered message %q: %w", underlyingMessage.FQMN(), err)
			}
			// If unimpacted, just use the same value.
			switch g.Options.FieldNameMode {
			case FieldNameModeJSON:
				result.Schema.Object.Properties[field.GetJsonName()] = modifiedSchema.Schema
			case FieldNameModeProto:
				result.Schema.Object.Properties[field.GetName()] = modifiedSchema.Schema
			}
			// TODO: deal with the dependencies here.
		}
	}

	return result, nil
}
