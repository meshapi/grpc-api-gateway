package integration

import (
	"context"

	"github.com/meshapi/grpc-api-gateway/examples/internal/gen/integration"
)

type PathParamsTestServer struct {
	integration.UnimplementedPathParamsTestServer
}

func (p PathParamsTestServer) Echo(
	ctx context.Context, req *integration.TestMessage) (*integration.TestMessage, error) {
	return req, nil
}
