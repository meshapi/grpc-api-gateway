package genopenapi

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
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

	schema, err := s.generator.openapiRegistry.getSchemaForMessage(location, fqmn)
	if err != nil {
		return fmt.Errorf("failed to render proto message %q to OpenAPI schema: %w", fqmn, err)
	}

	if s.Document.Object.Components.Object.Schemas == nil {
		s.Document.Object.Components.Object.Schemas = make(map[string]*openapiv3.Schema)
	}

	name, ok := s.generator.openapiRegistry.messageNames[fqmn]
	if !ok {
		return fmt.Errorf("unrecognized message: %s", fqmn)
	}

	s.includedSchemas[fqmn] = struct{}{}
	s.Document.Object.Components.Object.Schemas[name] = schema.Schema

	for _, dependency := range schema.Dependencies {
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

func (s *Session) includeEnum(location, fqen string) error {
	// TODO: we may need to do less object nesting here.

	// TODO: we may need to build the full FQMN here.
	// if schema is already added, skip processing it.
	if _, ok := s.includedSchemas[fqen]; ok {
		return nil
	}

	schema, err := s.generator.openapiRegistry.getSchemaForEnum(location, fqen)
	if err != nil {
		return fmt.Errorf("failed to render proto enum %q to OpenAPI schema: %w", fqen, err)
	}

	if s.Document.Object.Components.Object.Schemas == nil {
		s.Document.Object.Components.Object.Schemas = make(map[string]*openapiv3.Schema)
	}

	name, ok := s.generator.openapiRegistry.messageNames[fqen]
	if !ok {
		return fmt.Errorf("unrecognized message: %s", fqen)
	}

	s.includedSchemas[fqen] = struct{}{}
	s.Document.Object.Components.Object.Schemas[name] = schema

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
