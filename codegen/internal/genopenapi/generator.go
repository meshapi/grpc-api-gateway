package genopenapi

import (
	"fmt"
	"net/http"
	"strings"

	"dario.cat/mergo"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/protocomment"
)

type Generator struct {
	Options

	registry        *descriptor.Registry
	openapiRegistry *Registry
	commentRegistry *protocomment.Registry
}

func New(descriptorRegistry *descriptor.Registry, options Options) *Generator {
	commentRegistry := protocomment.NewRegistry(descriptorRegistry)
	return &Generator{
		Options:         options,
		registry:        descriptorRegistry,
		commentRegistry: commentRegistry,
		openapiRegistry: NewRegistry(&options, descriptorRegistry, commentRegistry),
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

			for _, service := range file.Services {
				if err := g.addServiceToSession(session, service); err != nil {
					return nil, fmt.Errorf("error generating OpenAPI definitions for service %q: %w", service.FQSN(), err)
				}
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
	if session.Document.Object.Paths == nil {
		session.Document.Object.Paths = make(map[string]*openapiv3.Path)
	}

	comments := g.openapiRegistry.commentRegistry.LookupService(service)
	//if comments != nil {
	//  operation.Description = r.renderComment(comments.Location)
	//}

	for _, method := range service.Methods {
		summary, description := "", ""

		if comments != nil && comments.Methods != nil {
			if methodComment := comments.Methods[int32(method.Index)]; methodComment != nil {
				result := renderComment(&g.Options, methodComment)
				firstParagraph := strings.Index(result, "\n\n")
				if firstParagraph > 0 {
					summary = result[:firstParagraph]
					description = result[firstParagraph+2:]
				} else {
					description = result
				}
			}
		}

		for _, binding := range method.Bindings {
			path := g.openapiRegistry.renderPath(binding)

			pathObject, exists := session.Document.Object.Paths[path]
			if !exists {
				pathObject = &openapiv3.Path{}
				session.Document.Object.Paths[path] = pathObject
			}

			var operationPtr **openapiv3.OperationCore

			switch binding.HTTPMethod {
			case http.MethodGet:
				operationPtr = &pathObject.Object.Get
			case http.MethodPost:
				operationPtr = &pathObject.Object.Post
			case http.MethodPut:
				operationPtr = &pathObject.Object.Put
			case http.MethodDelete:
				operationPtr = &pathObject.Object.Delete
			case http.MethodOptions:
				operationPtr = &pathObject.Object.Options
			case http.MethodHead:
				operationPtr = &pathObject.Object.Head
			case http.MethodPatch:
				operationPtr = &pathObject.Object.Patch
			case http.MethodTrace:
				operationPtr = &pathObject.Object.Trace
			default:
				continue
			}

			operation, err := g.renderOperation(binding)
			if err != nil {
				return fmt.Errorf("failed to render method %q: %w", method.FQMN(), err)
			}

			operation.Summary = summary
			operation.Description = description

			*operationPtr = operation
		}
	}
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
