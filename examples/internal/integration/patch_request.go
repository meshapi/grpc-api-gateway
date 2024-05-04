package integration

import (
	"context"

	"github.com/meshapi/grpc-api-gateway/examples/internal/gen/integration"
)

type PatchRequestTestServer struct {
	integration.UnimplementedPatchRequestTestServer
}

func (p PatchRequestTestServer) Patch(
	ctx context.Context, req *integration.PatchRequestSample) (*integration.PatchRequestSample, error) {
	return req, nil
}
