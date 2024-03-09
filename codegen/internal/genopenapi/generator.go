package genopenapi

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
)

type Generator struct {
	Options

	registry        *descriptor.Registry
	openapiRegistry *Registry

	// httpEndpointsMap is used to find duplicate HTTP specifications.
	//httpEndpointsMap map[endpointAnnotation]struct{}
}

func New(descriptorRegistry *descriptor.Registry, options Options) *Generator {
	return &Generator{
		Options:         options,
		registry:        descriptorRegistry,
		openapiRegistry: NewRegistry(&options),
	}
}

func (g *Generator) Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	var files []*descriptor.ResponseFile

	mergeOptions := []func(*mergo.Config){}

	if !g.Options.MergeWithOverwrite {
		mergeOptions = append(mergeOptions, mergo.WithAppendSlice)
	}

	if err := g.openapiRegistry.LoadFromDescriptorRegistry(g.registry); err != nil {
		return nil, err
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
			// TODO: if the document dose not have any schema or paths that was generated,
			// we should avoid generating the file.
			doc, ok := g.openapiRegistry.LookupDocument(file)
			if !ok {
				doc = &openapiv3.Extensible[openapiv3.DocumentCore]{}
			}

			// Merge with the root document if needed.
			if g.openapiRegistry.RootDocument != nil {
				if err := mergo.Merge(doc, g.openapiRegistry.RootDocument, mergeOptions...); err != nil {
					return nil, fmt.Errorf("failed to merge OpenAPI documents: %w", err)
				}
			}

			//err := g.addFileMessagesToDocument(doc, file)
			//if err != nil {
			//  return nil, fmt.Errorf("error generating OpenAPI for %q: %w", file.GetName(), err)
			//}

			file, err := g.writeDocument(file.GeneratedFilenamePrefix+".openapi", doc)
			if err != nil {
				return nil, fmt.Errorf("failed to write OpenAPI doc: %w", err)
			}

			files = append(files, file)
		}

	}

	return files, nil
}

func (g *Generator) addServiceToSession(doc *openapiv3.DocumentCore, service *descriptor.Service) error {
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

func (g *Generator) addFileMessagesToDocument(doc *openapiv3.DocumentCore, file *descriptor.File) error {
	return nil
}
