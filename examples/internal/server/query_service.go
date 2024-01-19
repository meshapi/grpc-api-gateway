package main

import (
	"context"

	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
)

type queryParamsTestServer struct {
	integration.UnimplementedQueryParamsTestServer
}

func (q queryParamsTestServer) Echo(
	ctx context.Context, req *integration.TestMessage) (*integration.TestMessage, error) {
	return req, nil
}
