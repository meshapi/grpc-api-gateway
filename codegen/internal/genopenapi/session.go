package genopenapi

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapi"
	"google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v3"
)

func (g *Generator) writeDocument(filePrefix string, doc *openapi.Document) (*descriptor.ResponseFile, error) {
	if doc == nil {
		return nil, nil
	}

	doc.OpenAPI = "3.1"
	if doc.Info == nil {
		doc.Info = &openapi.Info{
			Version: "version not set",
		}
	} else if doc.Info.Version == "" {
		doc.Info.Version = "version not set"
	}

	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI file: %w", err)
	}

	content := &bytes.Buffer{}
	var extension string

	switch g.OutputFormat {
	case OutputFormatYAML:
		if err := yaml.NewEncoder(content).Encode(doc); err != nil {
			return nil, fmt.Errorf("failed to marshal OpenAPI to yaml: %w", err)
		}
		extension = "yaml"
	case OutputFormatJSON:
		if err := json.NewEncoder(content).Encode(doc); err != nil {
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
