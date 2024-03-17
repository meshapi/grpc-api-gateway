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
	// httpEndpointsMap map[endpointAnnotation]struct{}
}

func New(descriptorRegistry *descriptor.Registry, options Options) *Generator {
	return &Generator{
		Options:         options,
		registry:        descriptorRegistry,
		openapiRegistry: NewRegistry(&options, descriptorRegistry),
	}
}

func (g *Generator) Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	var files []*descriptor.ResponseFile

	mergeOptions := []func(*mergo.Config){}

	if !g.Options.MergeWithOverwrite {
		mergeOptions = append(mergeOptions, mergo.WithAppendSlice)
	}

	if err := g.openapiRegistry.LoadFromDescriptorRegistry(); err != nil {
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
			doc := g.openapiRegistry.LookupDocument(file)
			if doc == nil {
				doc = &openapiv3.Document{}
			}

			err := g.addProtoMessageAndEnums(&doc.Object, file)
			if err != nil {
				return nil, fmt.Errorf("error generating OpenAPI for %q: %w", file.GetName(), err)
			}

			// Merge with the root document if needed.
			if g.openapiRegistry.RootDocument != nil {
				if err := mergo.Merge(doc, g.openapiRegistry.RootDocument, mergeOptions...); err != nil {
					return nil, fmt.Errorf("failed to merge OpenAPI documents: %w", err)
				}
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

func (g *Generator) addProtoMessageAndEnums(doc *openapiv3.DocumentCore, file *descriptor.File) error {
	if doc.Components == nil {
		doc.Components = &openapiv3.Components{}
	}

	for _, message := range file.Messages {
		fqmn := message.FQMN()
		schema, err := g.openapiRegistry.getSchemaForMessage(file.GetPackage(), fqmn)
		if err != nil {
			return fmt.Errorf("failed to render proto message %q to OpenAPI schema: %w", fqmn, err)
		}

		if doc.Components.Object.Schemas == nil {
			doc.Components.Object.Schemas = make(map[string]*openapiv3.Schema)
		}
		name := g.openapiRegistry.messageNames[fqmn]
		// TODO: we need to use a session here to include a metadata to the OpenAPI doc.
		// - for instance, adding to the schema should be a method to check for the object and all that.
		// TODO: handle the dependencies here as well.
		doc.Components.Object.Schemas[name] = schema.Schema
	}

	for _, enum := range file.Enums {
		fqen := enum.FQEN()
		schema, err := g.openapiRegistry.getSchemaForEnum(file.GetPackage(), fqen)
		if err != nil {
			return fmt.Errorf("failed to render proto enum %q to OpenAPI schema: %w", fqen, err)
		}

		if doc.Components.Object.Schemas == nil {
			doc.Components.Object.Schemas = make(map[string]*openapiv3.Schema)
		}
		name := g.openapiRegistry.messageNames[fqen]
		doc.Components.Object.Schemas[name] = schema
	}

	return nil
}
