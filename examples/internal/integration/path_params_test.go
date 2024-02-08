package integration_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
	"github.com/meshapi/grpc-rest-gateway/gateway"
)

func TestPathParams(t *testing.T) {
	manager := StartSharedTestServer()
	mux := gateway.NewServeMux()
	integration.RegisterPathParamsTestHandler(context.Background(), mux, manager.ClientConnection())

	tests := []struct {
		Name     string
		Request  *http.Request
		Response string
	}{
		{
			Name:     "Enum-Value",
			Request:  NewRequest("GET", "/path/enum/ID/1", nil, nil),
			Response: `{"id":"ID","priority":"Medium"}`,
		},
		{
			Name:     "Enum-Name",
			Request:  NewRequest("GET", "/path/enum/ID/Low", nil, nil),
			Response: `{"id":"ID","priority":"Low"}`,
		},
		{
			Name:     "RepeataedEnum-Value",
			Request:  NewRequest("GET", "/path/repeated-enum/ID/0,1", nil, nil),
			Response: `{"id":"ID","repeated_enums":["Zero","One"]}`,
		},
		{
			Name:     "RepeataedEnum-Name",
			Request:  NewRequest("GET", "/path/repeated-enum/ID/One,Two", nil, nil),
			Response: `{"id":"ID","repeated_enums":["One","Two"]}`,
		},
		{
			Name:     "OneOf",
			Request:  NewRequest("GET", "/path/oneof/ID/1/12", nil, nil),
			Response: `{"id":"ID","note_details":{"type_id":"1"},"month_num":12}`,
		},
		{
			Name:     "RepeatedStrings",
			Request:  NewRequest("GET", "/path/repeated_strings/ID/one,two,three", nil, nil),
			Response: `{"id":"ID","repeated_strings":["one","two","three"]}`,
		},
		{
			Name:     "CustomMethod-CatchAll",
			Request:  NewRequest("TEST", "/path/catch-all/blah/ID/something/interesting/with/slashes", nil, nil),
			Response: `{"id":"ID","nested_detail":{"text":"/something/interesting/with/slashes"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			AssertEchoRequest[*integration.TestMessage](t, mux, tt.Request, tt.Response)
		})
	}
}
