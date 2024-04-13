package internal

import "github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"

type FieldSchemaCustomization struct {
	Schema        *openapiv3.Schema
	PathParamName *string
	Required      bool
	ReadOnly      bool
	WriteOnly     bool
}
