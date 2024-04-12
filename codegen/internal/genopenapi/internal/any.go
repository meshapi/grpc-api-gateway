package internal

import "github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"

func AnySchema() OpenAPISchema {
	return OpenAPISchema{
		Schema: &openapiv3.Schema{
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
		},
	}
}
