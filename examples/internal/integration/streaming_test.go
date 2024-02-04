package integration_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
	"github.com/meshapi/grpc-rest-gateway/gateway"
)

func TestStreaming(t *testing.T) {
	manager := StartSharedTestServer()
	mux := gateway.NewServeMux()
	integration.RegisterStreamingTestHandler(context.Background(), mux, manager.ClientConnection())

	chunkedRequest := NewRequest(
		"POST", "/streaming/client/add", nil,
		NewChunkedBody(`{"value":10}`, `{"value":20}`, `{"value":30}`))

	tests := []struct {
		Name     string
		Request  *http.Request
		Response string
	}{
		{
			Name:     "ClientStreaming",
			Request:  chunkedRequest,
			Response: `{"sum":60,"count":3}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			AssertEchoRequest[*integration.AddResponse](t, mux, tt.Request, tt.Response)
		})
	}
}
