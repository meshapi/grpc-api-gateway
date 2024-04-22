package internal

import (
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
)

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

func MergeDefaultResponse(
	dest, src map[string]*openapiv3.Ref[openapiv3.Response]) map[string]*openapiv3.Ref[openapiv3.Response] {

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
