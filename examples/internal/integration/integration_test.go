package integration_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	integrationapi "github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
	"github.com/meshapi/grpc-rest-gateway/examples/internal/grpctest"
	"github.com/meshapi/grpc-rest-gateway/examples/internal/integration"
	"github.com/meshapi/grpc-rest-gateway/gateway"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"
)

var (
	// StartSharedTestServer initializes and starts the global server that is shared among all tests.
	//
	// this server should not be shutdown until all tests are spun down. this instance does NOT have the
	// *testing.T installed.
	//
	// NOTE: StartSharedTestServer is idempotent and can be called repeatedly.
	StartSharedTestServer = sync.OnceValue(func() *grpctest.Manager {
		instance := grpctest.NewManager(nil, "shared-grpc-server", func(s *grpc.Server) {
			// register all services here.
			integrationapi.RegisterQueryParamsTestServer(s, &integration.QueryParamsTestServer{})
			integrationapi.RegisterPathParamsTestServer(s, &integration.PathParamsTestServer{})
			integrationapi.RegisterPatchRequestTestServer(s, &integration.PatchRequestTestServer{})
			integrationapi.RegisterStreamingTestServer(s, &integration.StreamingTestServer{})
		})
		instance.Start()

		return instance
	})
)

// NewRequest is a shortcut method for creating HTTP requests with query parameters.
func NewRequest(method, path string, values url.Values, body io.Reader) *http.Request {
	if values != nil {
		path = path + "?" + values.Encode()
	}

	return httptest.NewRequest(method, path, body)
}

// Unmarshal unmarshals a proto message from a reader instance.
func Unmarshal(t *testing.T, reader io.Reader, value protoreflect.ProtoMessage) bool {
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("failed to read: %s", err)
		return false
	}

	if err := protojson.Unmarshal(content, value); err != nil {
		t.Errorf("failed to unmarshal: %s", err)
		return false
	}

	return true
}

// NewChunkedBody returns a reader that simulates a chunked request for a client streaming request.
func NewChunkedBody(chunks ...string) io.Reader {
	buffer := &bytes.Buffer{}

	for _, chunk := range chunks {
		buffer.WriteString(chunk)
		buffer.Write([]byte("\r\n"))
	}

	return buffer
}

// AssertEchoRequest runs the mux's handler for the given HTTP request and asserts the response matches the expected
// JSON text.
func AssertEchoRequest[T protoreflect.ProtoMessage](t *testing.T, mux *gateway.ServeMux, req *http.Request, responseText string) {
	responseRecorder := httptest.NewRecorder()
	mux.ServeHTTP(responseRecorder, req)

	if responseRecorder.Result().StatusCode != 200 {
		t.Fatalf("received status code %d", responseRecorder.Result().StatusCode)
		return
	}

	var zeroMessage T

	expectedResponse := reflect.New(reflect.TypeOf(zeroMessage).Elem()).Interface().(T)
	if !Unmarshal(t, strings.NewReader(responseText), expectedResponse) {
		return
	}

	body := responseRecorder.Result().Body
	defer body.Close()

	response := reflect.New(reflect.TypeOf(zeroMessage).Elem()).Interface().(T)
	if !Unmarshal(t, body, response) {
		return
	}

	if diff := cmp.Diff(expectedResponse, response, protocmp.Transform()); diff != "" {
		t.Fatalf("incorrect response:\n%s", diff)
		return
	}
}

// AssertChunkedResponse runs the mux's handler for the given HTTP request and asserts the responses matches the
// expected JSON text chunks.
func AssertChunkedResponse[T protoreflect.ProtoMessage](t *testing.T, mux *gateway.ServeMux, req *http.Request, responseTexts []string) {

	responseRecorder := httptest.NewRecorder()
	mux.ServeHTTP(responseRecorder, req)

	if responseRecorder.Result().StatusCode != 200 {
		t.Fatalf("received status code %d", responseRecorder.Result().StatusCode)
		return
	}

	var zeroMessage T

	for _, expectedResponseText := range responseTexts {
		expectedResponse := reflect.New(reflect.TypeOf(zeroMessage).Elem()).Interface().(T)
		receivedResponse := reflect.New(reflect.TypeOf(zeroMessage).Elem()).Interface().(T)

		line, err := responseRecorder.Body.ReadString('\n')
		if err != nil {
			t.Fatalf("error reading chunk: %s", err)
			return
		}

		if !Unmarshal(t, strings.NewReader(expectedResponseText), expectedResponse) {
			return
		}

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
}
