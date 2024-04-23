package genopenapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/internal"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/protocomment"
)

type Generator struct {
	Options

	registry        *descriptor.Registry
	commentRegistry *protocomment.Registry

	// rootDocument is the global and top-level document loaded from the global config.
	rootDocument internal.OpenAPIDocument
	// files are the mapped files from for each proto file.
	files map[*descriptor.File]internal.OpenAPIDocument
	// schemaNames holds a one to one association of message/enum FQNs to the generated OpenAPI schema name.
	schemaNames map[string]string
	// messages holds a map of message FQMNs to parsed configuration.
	messages map[string]*internal.OpenAPIMessageSpec
	// enums holds a map of enum FQENs to parsed configuration.
	enums map[string]*internal.OpenAPIEnumSpec
	// schemas are already processed schemas that can be readily used.
	schemas map[string]internal.OpenAPISchema
	// services holds a reference to the service and any matched configuration for it.
	services map[string]*internal.OpenAPIServiceSpec
}

func New(descriptorRegistry *descriptor.Registry, options Options) *Generator {
	commentRegistry := protocomment.NewRegistry(descriptorRegistry)
	return &Generator{
		Options:         options,
		registry:        descriptorRegistry,
		commentRegistry: commentRegistry,
		rootDocument:    internal.OpenAPIDocument{},
		files:           map[*descriptor.File]internal.OpenAPIDocument{},
		schemaNames:     map[string]string{},
		messages:        map[string]*internal.OpenAPIMessageSpec{},
		enums:           map[string]*internal.OpenAPIEnumSpec{},
		schemas: map[string]internal.OpenAPISchema{
			fqmnAny: internal.AnySchema(),
		},
		services: map[string]*internal.OpenAPIServiceSpec{},
	}
}

func (g *Generator) Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	var files []*descriptor.ResponseFile

	if err := g.loadFromDescriptorRegistry(); err != nil {
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
			doc := g.LookupFile(file)
			if doc.Document == nil {
				doc.Document = &openapiv3.Document{}
			}

			session := g.newSession(doc.Document)

			err := session.addMessageAndEnums(file)
			if err != nil {
				return nil, fmt.Errorf("error generating OpenAPI for %q: %w", file.GetName(), err)
			}

			for _, service := range file.Services {
				if err := session.addService(service); err != nil {
					return nil, fmt.Errorf("error generating OpenAPI definitions for service %q: %w", service.FQSN(), err)
				}
			}

			// Merge with the root document if needed.
			if g.rootDocument.Document != nil {
				if err := g.mergeObjects(doc.Document, g.rootDocument.Document); err != nil {
					return nil, fmt.Errorf("failed to merge OpenAPI documents: %w", err)
				}
			}

			file, err := g.writeDocument(file.GeneratedFilenamePrefix+openAPIOutputSuffix, doc.Document)
			if err != nil {
				return nil, fmt.Errorf("failed to write OpenAPI doc: %w", err)
			}

			files = append(files, file)
		}

	}

	return files, nil
}

func (s *Session) addService(service *descriptor.Service) error {
	if s.Document.Object.Paths == nil {
		s.Document.Object.Paths = make(map[string]*openapiv3.Path)
	}

	comments := s.commentRegistry.LookupService(service)
	defaultResponses, err := s.defaultResponsesForService(service)
	if err != nil {
		return err
	}

	for _, method := range service.Methods {
		summary, description := "", ""

		if comments != nil && comments.Methods != nil {
			if methodComment := comments.Methods[int32(method.Index)]; methodComment != nil {
				result := s.renderComment(methodComment)
				firstParagraph := strings.Index(result, "\n\n")
				if firstParagraph > 0 {
					summary = result[:firstParagraph]
					description = result[firstParagraph+2:]
				} else {
					description = result
				}
			}
		}

		var pathParamAliasMap map[string]string
		customizedOperation, err := s.getCustomizedMethodOperation(method)
		if err != nil {
			return fmt.Errorf("failed to map method configs to OpenAPI operation object for %q: %w", method.FQMN(), err)
		}

		for _, binding := range method.Bindings {
			pathParamAliasMap = s.updatePathParameterAliasesMap(pathParamAliasMap, binding)

			path := renderPath(binding, pathParamAliasMap)

			pathObject, exists := s.Document.Object.Paths[path]
			if !exists {
				pathObject = &openapiv3.Path{}
			}

			var operationPtr **openapiv3.Operation

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

			if !exists {
				// add this only when the operation is using an HTTP endpoint that can be recorded in
				// the OpenAPI document.
				s.Document.Object.Paths[path] = pathObject
			}

			operation, err := s.renderOperation(binding, defaultResponses)
			if err != nil {
				return fmt.Errorf("failed to render method %q: %w", method.FQMN(), err)
			}

			operation.Object.Summary = summary
			operation.Object.Description = description

			if customizedOperation != nil {
				if err := s.mergeObjects(customizedOperation, operation); err != nil {
					return err
				}

				*operationPtr = customizedOperation
			} else {
				*operationPtr = operation
			}

		}
	}
	return nil
}

func (s *Session) addMessageAndEnums(file *descriptor.File) error {
	if s.Document.Object.Components == nil {
		s.Document.Object.Components = &openapiv3.Components{}
	}

	for _, message := range file.Messages {
		// Skip all messages that are generated to be used as MapEntry objects.
		if message.IsMapEntry() {
			continue
		}

		if err := s.includeMessage(message.FQMN()); err != nil {
			return fmt.Errorf("failed to process message %q: %w", message.FQMN(), err)
		}
	}

	for _, enum := range file.Enums {
		if err := s.includeEnum(enum.FQEN()); err != nil {
			return fmt.Errorf("failed to process enum %q: %w", enum.FQEN(), err)
		}
	}

	return nil
}
