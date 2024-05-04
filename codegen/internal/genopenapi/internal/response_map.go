package internal

import (
	"fmt"

	"github.com/meshapi/grpc-api-gateway/api/openapi"
	"github.com/meshapi/grpc-api-gateway/codegen/internal/genopenapi/openapimap"
	"github.com/meshapi/grpc-api-gateway/codegen/internal/openapiv3"
)

type DefaultResponse struct {
	Response     *openapiv3.Ref[openapiv3.Response]
	Dependencies SchemaDependencyStore
	Processed    bool
}

type DefaultResponses map[string]*DefaultResponse

// MapDefaultResponses is similar to openapimap.ResponseMap but it uses DefaultResponse type.
func MapDefaultResponses(responses map[string]*openapi.Response) (DefaultResponses, error) {
	if responses == nil {
		return nil, nil
	}

	result := make(DefaultResponses, len(responses))
	for key, responseFromProto := range responses {
		response, err := openapimap.Response(responseFromProto)
		if err != nil {
			return nil, fmt.Errorf("invalid response for %q: %w", key, err)
		}
		result[key] = &DefaultResponse{
			Response:     response,
			Dependencies: nil,
			Processed:    false,
		}
	}

	return result, nil
}

func MergeDefaultResponseSpec(dest, src map[string]*openapi.Response) map[string]*openapi.Response {
	if dest == nil {
		return src
	} else if src == nil {
		return dest
	}

	for key, value := range src {
		dest[key] = value
	}

	return dest
}

func MergeDefaultResponses(dest, src DefaultResponses) DefaultResponses {
	if dest == nil {
		return src
	} else if src == nil {
		return dest
	}

	for key, value := range src {
		if _, exists := dest[key]; exists {
			continue
		}
		dest[key] = value
	}

	return dest
}
