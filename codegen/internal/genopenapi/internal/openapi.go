package internal

import (
	"github.com/meshapi/grpc-api-gateway/api"
	"github.com/meshapi/grpc-api-gateway/codegen/internal/openapiv3"
)

// OpenAPISpec is a wrapper around *api.OpenAPISpec with additional context of filename.
type OpenAPISpec struct {
	*api.OpenAPISpec

	// Filename is the config file name that is the source of this spec.
	Filename string
}

type SourceInfo struct {
	Filename     string
	ProtoPackage string
}

type OpenAPIMessageSpec struct {
	*api.OpenAPIMessageSpec
	SourceInfo
}

type OpenAPIServiceSpec struct {
	*api.OpenAPIServiceSpec
	SourceInfo
}

type OpenAPIEnumSpec struct {
	*api.OpenAPIEnumSpec
	SourceInfo
}

// OpenAPISchema is a wrapper around an openapiv3.Schema with a dependency store.
type OpenAPISchema struct {
	// Schema is the already mapped and processed OpenAPI schema for a proto message/enum.
	Schema *openapiv3.Schema
	// Dependencies is the list of enum or proto message dependencies that must be included in the same
	// OpenAPI document.
	Dependencies SchemaDependencyStore
}

// OpenAPIDocument is a wrapper around an openapiv3.Document with document generation configs.
type OpenAPIDocument struct {
	// Document is the already mapped and processed OpenAPI document for a proto file.
	Document *openapiv3.Document
	// DefaultResponses are the already mapped default responses for all operations for this document/service.
	DefaultResponses DefaultResponses
}
