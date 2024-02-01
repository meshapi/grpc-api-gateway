package integration_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		})
		instance.Start()

		return instance
	})
)

func NewRequest(method, path string, values url.Values, body io.Reader) *http.Request {
	if values != nil {
		path = path + "?" + values.Encode()
	}

	return httptest.NewRequest(method, path, body)
}

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

func AssertEchoRequest(t *testing.T, mux *gateway.ServeMux, req *http.Request, responseText string) {
	responseRecorder := httptest.NewRecorder()
	mux.ServeHTTP(responseRecorder, req)

	if responseRecorder.Result().StatusCode != 200 {
		t.Fatalf("received status code %d", responseRecorder.Result().StatusCode)
		return
	}

	expectedResponse := &integrationapi.TestMessage{}
	if !Unmarshal(t, strings.NewReader(responseText), expectedResponse) {
		return
	}

	body := responseRecorder.Result().Body
	defer body.Close()

	response := &integrationapi.TestMessage{}
	if !Unmarshal(t, body, response) {
		return
	}

	if diff := cmp.Diff(expectedResponse, response, protocmp.Transform()); diff != "" {
		t.Fatalf("incorrect response:\n%s", diff)
		return
	}
}
