package main

import (
	"context"

	"github.com/meshapi/grpc-rest-gateway/examples/internal/proto/echo"
)

type EchoService struct {
	echo.UnimplementedEchoServiceServer
}

func (e EchoService) Echo(ctx context.Context, req *echo.SimpleMessage) (*echo.SimpleMessage, error) {
	return req, nil
}
