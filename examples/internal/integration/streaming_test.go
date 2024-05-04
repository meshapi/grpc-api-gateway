package integration_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/meshapi/grpc-api-gateway/examples/internal/gen/integration"
	"github.com/meshapi/grpc-api-gateway/gateway"
)

func TestClientStreamingChunked(t *testing.T) {
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

func TestServerStreamingChunked(t *testing.T) {
	manager := StartSharedTestServer()
	mux := gateway.NewServeMux()
	integration.RegisterStreamingTestHandler(context.Background(), mux, manager.ClientConnection())

	tests := []struct {
		Name      string
		Request   *http.Request
		Responses []string
	}{
		{
			Name:    "Generate",
			Request: NewRequest("GET", "/streaming/server/generate", url.Values{"count": []string{"3"}}, nil),
			Responses: []string{
				`{"value": 1, "index": 0}`,
				`{"value": 3, "index": 1}`,
				`{"value": 7, "index": 2}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			AssertChunkedResponse[*integration.GenerateResponse](t, mux, tt.Request, tt.Responses)
		})
	}
}

func TestServerAndClientStreamingChunked(t *testing.T) {
	manager := StartSharedTestServer()
	mux := gateway.NewServeMux()
	integration.RegisterStreamingTestHandler(context.Background(), mux, manager.ClientConnection())

	tests := []struct {
		Name      string
		Request   *http.Request
		Responses []string
	}{
		{
			Name:    "Generate",
			Request: NewRequest("POST", "/streaming/bidi/bulk-capitalize", nil, NewChunkedBody(`{"text":"hi"}`, `{"text":"bye"}`)),
			Responses: []string{
				`{"text": "HI","index":1}`,
				`{"text": "BYE","index":2}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			AssertChunkedResponse[*integration.BulkCapitalizeResponse](t, mux, tt.Request, tt.Responses)
		})
	}
}
