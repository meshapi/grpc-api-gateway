package genopenapi

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/internal"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/openapimap"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/pathfilter"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"gopkg.in/yaml.v3"
)

type Session struct {
	*Generator
	Document *openapiv3.Document

	includedSchemas map[string]struct{}
}

func (g *Generator) newSession(doc *openapiv3.Document) *Session {
	return &Session{
		Document:        doc,
		Generator:       g,
		includedSchemas: make(map[string]struct{}),
	}
}

func (s *Session) includeMessage(fqmn string) error {
	if _, ok := s.includedSchemas[fqmn]; ok {
		return nil
	}

	schema, err := s.getSchemaForMessage("", fqmn)
	if err != nil {
		return fmt.Errorf("failed to render proto message %q to OpenAPI schema: %w", fqmn, err)
	}

	schemas := s.Document.Object.Components.Object.Schemas
	if schemas == nil {
		schemas = make(map[string]*openapiv3.Schema)
		s.Document.Object.Components.Object.Schemas = schemas
	}

	name, ok := s.schemaNames[fqmn]
	if !ok {
		return fmt.Errorf("unrecognized message: %s", fqmn)
	}

	s.includedSchemas[fqmn] = struct{}{}
	schemas[name] = schema.Schema

	if err := s.includeDependencies(schema.Dependencies); err != nil {
		return err
	}

	return nil
}

func (s *Session) includeEnum(fqen string) error {
	if _, ok := s.includedSchemas[fqen]; ok {
		return nil
	}

	schema, err := s.getSchemaForEnum("", fqen)
	if err != nil {
		return fmt.Errorf("failed to render proto enum %q to OpenAPI schema: %w", fqen, err)
	}

	schemas := s.Document.Object.Components.Object.Schemas
	if schemas == nil {
		schemas = make(map[string]*openapiv3.Schema)
		s.Document.Object.Components.Object.Schemas = schemas
	}

	name, ok := s.schemaNames[fqen]
	if !ok {
		return fmt.Errorf("unrecognized message: %s", fqen)
	}

	s.includedSchemas[fqen] = struct{}{}
	schemas[name] = schema

	return nil
}

func (s *Session) includeDependency(location string, dependency internal.SchemaDependency) error {
	switch dependency.Kind {
	case internal.DependencyKindMessage:
		return s.includeMessage(dependency.FQN)
	case internal.DependencyKindEnum:
		return s.includeEnum(dependency.FQN)
	}
	return fmt.Errorf("unexpected dependency kind: %v", dependency.Kind)
}

func (s *Session) includeDependencies(dependencies internal.SchemaDependencyStore) error {
	for fqn, dependency := range dependencies {
		switch dependency.Kind {
		case internal.DependencyKindMessage:
			if err := s.includeMessage(fqn); err != nil {
				return fmt.Errorf("failed to include message dependency %q: %w", fqn, err)
			}
		case internal.DependencyKindEnum:
			if err := s.includeEnum(fqn); err != nil {
				return fmt.Errorf("failed to include enum dependency %q: %w", fqn, err)
			}
		}
	}

	return nil
}

func (g *Generator) defaultResponsesForService(
	service *descriptor.Service) (map[string]*openapiv3.Ref[openapiv3.Response], error) {

	var fromConfig, fromProto map[string]*openapi.Response
	if doc := g.services[service.FQSN()]; doc != nil {
		fromConfig = doc.GetDocument().GetConfig().GetDefaultResponses()
	}
	serviceSpec, ok := proto.GetExtension(service.Options, api.E_OpenapiServiceDoc).(*openapi.Document)
	if ok && serviceSpec != nil {
		fromProto = serviceSpec.GetConfig().GetDefaultResponses()
	}
	serviceDefaultResponseSpec := internal.MergeDefaultResponseSpec(fromProto, fromConfig)

	if serviceDefaultResponseSpec == nil {
		return g.LookupFile(service.File).DefaultResponses, nil
	}

	responses, err := openapimap.ResponseMap(serviceDefaultResponseSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to map default responses to OpenAPI responses: %w", err)
	}

	return internal.MergeDefaultResponse(responses, g.LookupFile(service.File).DefaultResponses), nil
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
func (s *Session) renderMessageSchemaWithFilter(
	message *descriptor.Message, filter *pathfilter.Instance) (internal.OpenAPISchema, error) {

	originalSchema, err := s.getSchemaForMessage("", message.FQMN())
	if err != nil {
		return internal.OpenAPISchema{}, fmt.Errorf("failed to get schema for %q: %w", message.FQMN(), err)
	}

	schemaCopy := *originalSchema.Schema
	schemaCopy.Object.Properties = make(map[string]*openapiv3.Schema)
	result := internal.OpenAPISchema{
		Schema:       &schemaCopy,
		Dependencies: originalSchema.Dependencies.Copy(),
	}

	hasRequiredSlice := len(schemaCopy.Object.Required) > 0
	var fieldsToRemove map[string]struct{}

	for _, field := range message.Fields {
		impacted, instance := filter.HasString(field.GetName())

		switch {
		case !impacted:
			// If unimpacted, just use the same value.
			fieldName := s.fieldName(field)
			result.Schema.Object.Properties[fieldName] = originalSchema.Schema.Object.Properties[fieldName]
		case instance.Excluded:
			// path parameter cannot be excluded and be a non-scalar type so there is only one scenario in which dependencies
			// need to get updated, if the type is an enum type.
			if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
				enum, err := s.registry.LookupEnum(message.FQMN(), field.GetTypeName())
				if err != nil {
					return internal.OpenAPISchema{}, fmt.Errorf("failed to find enum %q: %w", field.GetTypeName(), err)
				}
				result.Dependencies.Drop(enum.FQEN())
			}
			if hasRequiredSlice {
				fieldsToRemove = s.appendFieldToRemovedFields(fieldsToRemove, field)
			}
		default:
			underlyingMessage, err := s.registry.LookupMessage(message.FQMN(), field.GetTypeName())
			if err != nil {
				return internal.OpenAPISchema{}, fmt.Errorf("failed to find message %q: %w", field.GetTypeName(), err)
			}
			modifiedSchema, err := s.renderMessageSchemaWithFilter(underlyingMessage, instance)
			if err != nil {
				return internal.OpenAPISchema{}, fmt.Errorf("failed to render filtered message %q: %w", underlyingMessage.FQMN(), err)
			}
			if len(modifiedSchema.Schema.Object.Properties) != 0 {
				result.Schema.Object.Properties[s.fieldName(field)] = modifiedSchema.Schema
				result.Dependencies.Drop(underlyingMessage.FQMN())
			} else if hasRequiredSlice {
				// NOTE: Messages that are empty are considered to be deliberately empty so they do not get removed.
				// However, the schemas that become empty as part of this filtering down get removed.
				fieldsToRemove = s.appendFieldToRemovedFields(fieldsToRemove, field)
			}
		}
	}

	if fieldsToRemove != nil {
		schemaCopy.Object.Required = internal.FilteredStringSlice(schemaCopy.Object.Required, fieldsToRemove)
	}

	return result, nil
}

func (s *Session) appendFieldToRemovedFields(table map[string]struct{}, field *descriptor.Field) map[string]struct{} {
	if table == nil {
		return map[string]struct{}{s.fieldName(field): {}}
	}

	table[s.fieldName(field)] = struct{}{}
	return table
}

func (s *Session) getCustomizedMethodOperation(method *descriptor.Method) (*openapiv3.Operation, error) {
	var operation *openapiv3.Operation

	if serviceConfig, ok := s.services[method.Service.FQSN()]; ok && serviceConfig.Methods != nil {
		methodConfig, ok := serviceConfig.Methods[method.GetName()]
		if ok {
			result, err := openapimap.Operation(methodConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to map operation: %w", err)
			}

			operation = result
		}
	}

	protoConfig, ok := proto.GetExtension(method.Options, api.E_OpenapiOperation).(*openapi.Operation)
	if ok && protoConfig != nil {
		result, err := openapimap.Operation(protoConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to map operation from proto config: %w", err)
		}

		if operation == nil {
			operation = result
		} else if err := s.mergeObjects(operation, result); err != nil {
			return nil, err
		}
	}

	return operation, nil
}
