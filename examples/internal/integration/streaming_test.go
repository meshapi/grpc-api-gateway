package integration_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
	"github.com/meshapi/grpc-rest-gateway/gateway"
	"google.golang.org/protobuf/testing/protocmp"
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
			responseRecorder := httptest.NewRecorder()
			mux.ServeHTTP(responseRecorder, tt.Request)

			if responseRecorder.Result().StatusCode != 200 {
				t.Fatalf("received status code %d", responseRecorder.Result().StatusCode)
				return
			}

			expectedResponse := &integration.GenerateResponse{}
			receivedResponse := &integration.GenerateResponse{}
			for _, expectedResponseText := range tt.Responses {
				line, err := responseRecorder.Body.ReadString('\n')
				if err != nil {
					t.Fatalf("error reading chunk: %s", err)
					return
				}
				expectedResponse.Reset()
				if !Unmarshal(t, strings.NewReader(expectedResponseText), expectedResponse) {
					return
				}

				receivedResponse.Reset()
				if !Unmarshal(t, strings.NewReader(line), receivedResponse) {
					return
				}

				if diff := cmp.Diff(expectedResponse, receivedResponse, protocmp.Transform()); diff != "" {
					t.Fatalf("incorrect response:\n%s", diff)
					return
				}
			}

			if line, err := responseRecorder.Body.ReadString('\n'); err != io.EOF {
				t.Fatalf("expected EOF but received %q", line)
			}
		})
	}
}
