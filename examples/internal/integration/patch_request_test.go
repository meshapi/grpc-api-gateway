package integration_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
	"github.com/meshapi/grpc-rest-gateway/gateway"
)

func TestPatchRequest(t *testing.T) {
	manager := StartSharedTestServer()
	mux := gateway.NewServeMux()
	integration.RegisterPatchRequestTestHandler(context.Background(), mux, manager.ClientConnection())

	tests := []struct {
		Name     string
		Request  *http.Request
		Response string
	}{
		{
			Name: "PUT-Full",
			Request: NewRequest(
				"PUT", "/patch/full/ID", nil, strings.NewReader(
					`{"body":{"note_content":"something","priority":"Medium"}}`)),
			Response: `{"body":{"id":"ID","note_content":"something","priority":"Medium"}}`,
		},
		{
			Name: "PUT-Full-WithMask",
			Request: NewRequest(
				"PUT", "/patch/full/ID", nil, strings.NewReader(
					`{"body":{"note_content":"something","priority":"Medium"},"update_mask":"noteContent"}`)),
			Response: `{"body":{"id":"ID","note_content":"something","priority":"Medium"},"update_mask":"noteContent"}`,
		},
		{
			Name: "PATCH-Full",
			Request: NewRequest(
				"PATCH", "/patch/full/ID", nil, strings.NewReader(
					`{"body":{"note_content":"something","priority":"Medium"}}`)),
			Response: `{"body":{"id":"ID","note_content":"something","priority":"Medium"}}`,
		},
		{
			Name: "PUT-Body",
			Request: NewRequest(
				"PUT", "/patch/body/ID", nil, strings.NewReader(
					`{"note_content":"something","priority":"Medium"}`)),
			Response: `{"body":{"id":"ID","note_content":"something","priority":"Medium"}}`,
		},
		{
			Name: "PATCH-Body",
			Request: NewRequest(
				"PATCH", "/patch/body/ID", nil, strings.NewReader(
					`{"note_content":"something","priority":"Medium"}`)),
			Response: `{"body":{"id":"ID","note_content":"something","priority":"Medium"},"update_mask":"noteContent,priority"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			AssertEchoRequest[*integration.PatchRequestSample](t, mux, tt.Request, tt.Response)
		})
	}
}
