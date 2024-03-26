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

			session := g.newSession(doc)

			err := g.addProtoMessageAndEnums(session, file)
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

func (g *Generator) addServiceToSession(session *Session, service *descriptor.Service) error {
	return nil
}

// func (g *Generator) addProtoMessageAndEnums(doc *openapiv3.DocumentCore, file *descriptor.File) error {
func (g *Generator) addProtoMessageAndEnums(session *Session, file *descriptor.File) error {
	if session.Document.Object.Components == nil {
		session.Document.Object.Components = &openapiv3.Components{}
	}

	for _, message := range file.Messages {
		// Skip all messages that are generated to be used as MapEntry objects.
		if message.IsMapEntry() {
			continue
		}

		if err := session.includeMessage(file.GetPackage(), message.FQMN()); err != nil {
			return fmt.Errorf("failed to process message %q: %w", message.FQMN(), err)
		}
	}

	for _, enum := range file.Enums {
		if err := session.includeEnum(file.GetPackage(), enum.FQEN()); err != nil {
			return fmt.Errorf("failed to process enum %q: %w", enum.FQEN(), err)
		}
	}

	return nil
}
