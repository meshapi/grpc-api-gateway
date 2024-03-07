package genopenapi

import (
	"fmt"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapi"
)

type Generator struct {
	Options

	registry        *descriptor.Registry
	openapiRegistry *openapi.Registry

	// httpEndpointsMap is used to find duplicate HTTP specifications.
	//httpEndpointsMap map[endpointAnnotation]struct{}
}

func New(descriptorRegistry *descriptor.Registry, options Options) *Generator {
	return &Generator{
		Options:         options,
		registry:        descriptorRegistry,
		openapiRegistry: openapi.NewRegistry(),
	}
}

func (g *Generator) Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	var files []*descriptor.ResponseFile

	mergeOptions := []func(*mergo.Config){}

	if g.Options.AllowDeleteBody {
		mergeOptions = append(mergeOptions, mergo.WithAppendSlice)
	}

	// we need to first prepare the registry and preload all schemas
	// then we load the root one and prepare a global session.
	if g.Options.GlobalOpenAPIConfigFile != "" {
		globalConfigFilePath := filepath.Join(g.ConfigSearchPath, g.Options.GlobalOpenAPIConfigFile)
		if err := g.openapiRegistry.LoadGlobalConfig(globalConfigFilePath); err != nil {
			return nil, fmt.Errorf("failed to read global openapi file %q: %w", g.Options.GlobalOpenAPIConfigFile, err)
		}
	}

	for _, file := range targets {
		// we need to create a session based on the super session and then pass it down.
		// for now, for simplicity, we won't create this session and instead directly pass it down.
		// var doc *openapi.Document

		switch g.OutputMode {
		case OutputModeMerge:
			//doc = g.openapiRegistry.RootDocument
		case OutputModePerService:
			// loop over the services and add them one by one.
		case OutputModePerProtoFile:
			// we need to create a new session here but we'll just cheat here for a second.
			// TODO: if the document dose not have any schema or paths, we should avoid generating the file.
			doc, ok := g.openapiRegistry.LookupDocument(file)
			if !ok {
				doc = &openapi.Document{}
			}

			// Merge with the root document if needed.
			if g.openapiRegistry.RootDocument != nil {
				if err := mergo.Merge(doc, g.openapiRegistry.RootDocument, mergeOptions...); err != nil {
					return nil, fmt.Errorf("failed to merge OpenAPI documents: %w", err)
				}
			}

			err := g.addFileMessagesToDocument(doc, file)
			if err != nil {
				return nil, fmt.Errorf("error generating OpenAPI for %q: %w", file.GetName(), err)
			}

			file, err := g.writeDocument(file.GeneratedFilenamePrefix+".openapi", doc)
			if err != nil {
				return nil, fmt.Errorf("failed to write OpenAPI doc: %w", err)
			}

			files = append(files, file)
		}

	}

	return files, nil
}

func (g *Generator) addServiceToSession(doc *openapi.Document, service *descriptor.Service) error {
	// we would in theory merge them but here we just want to prepare the files.

	// create a session here.
	//session := &Session{
	//  generator: g,
	//  document:  g.openapiRegistry.RootDocument,
	//}

	//if err := session.WriteDocument(); err != nil {
	//  return nil, fmt.Errorf("failed to write OpenAPI document: %w", err)
	//}

	return nil
}

func (g *Generator) addFileMessagesToDocument(doc *openapi.Document, file *descriptor.File) error {
	return nil
}
