package internal

import "github.com/meshapi/grpc-api-gateway/codegen/internal/openapiv3"

type ErrorResponse struct {
	Response *openapiv3.Ref[openapiv3.Response]

	// ReferenceIsResolved indicated whether or not the RPC error status is set on the resposne object.
	ReferenceIsResolved bool
}

var (
	anySchema      OpenAPISchema
	httpBodySchema OpenAPISchema
	errorResponse  ErrorResponse
)

func AnySchema() OpenAPISchema {
	if anySchema.Schema == nil {
		anySchema.Schema = &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:        openapiv3.TypeSet{openapiv3.TypeObject},
				Description: "Any contains an arbitrary schema along with a URL to help identify the type of the schema.",
				Properties: map[string]*openapiv3.Schema{
					"@type": {
						Object: openapiv3.SchemaCore{
							Type:        openapiv3.TypeSet{openapiv3.TypeString},
							Description: "A URL/resource name that uniquely identifies the type of the schema.",
						},
					},
				},
			},
		}
	}

	return anySchema
}

func HTTPBodySchema() OpenAPISchema {
	if httpBodySchema.Schema == nil {
		httpBodySchema.Schema = &openapiv3.Schema{
			Object: openapiv3.SchemaCore{
				Type:   openapiv3.TypeSet{openapiv3.TypeString},
				Format: "byte",
			},
		}
	}

	return httpBodySchema
}

func DefaultErrorResponse() ErrorResponse {
	if errorResponse.Response == nil {
		errorResponse.Response = &openapiv3.Ref[openapiv3.Response]{
			Data: openapiv3.Response{
				Object: openapiv3.ResponseCore{
					Description: "An unexpected error response.",
					Content: map[string]*openapiv3.MediaType{
						"application/json": {
							Object: openapiv3.MediaTypeCore{
								Schema: &openapiv3.Schema{},
							},
						},
					},
				},
			},
		}
	}

	return errorResponse
}

func SetErrorResponseRef(ref string) {
	// NOTE: Love the code below lol
	errorResponse.Response.Data.Object.Content["application/json"].Object.Schema.Object.Ref = ref
	errorResponse.ReferenceIsResolved = true
}
