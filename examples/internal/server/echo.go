package main

import (
	"context"

	"github.com/meshapi/grpc-api-gateway/examples/internal/gen/echo"
)

type EchoService struct {
	echo.UnimplementedEchoServiceServer
}

func (e EchoService) Echo(ctx context.Context, req *echo.SimpleMessage) (*echo.SimpleMessage, error) {
	return req, nil
}
