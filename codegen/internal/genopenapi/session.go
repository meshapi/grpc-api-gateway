package genopenapi

import "github.com/meshapi/grpc-rest-gateway/codegen/internal/openapi"

// Session is an OpenAPI document session and will be used to generate a single OpenAPI documentation file.
type Session struct {
	document openapi.Document
}
