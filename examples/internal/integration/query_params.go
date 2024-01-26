package integration

import (
	"context"

	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
)

type QueryParamsTestServer struct {
	integration.UnimplementedQueryParamsTestServer
}

func (q QueryParamsTestServer) Echo(
	ctx context.Context, req *integration.TestMessage) (*integration.TestMessage, error) {
	return req, nil
}
