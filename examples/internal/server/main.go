package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/echo"
	integrationapi "github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
	"github.com/meshapi/grpc-rest-gateway/examples/internal/integration"
	"github.com/meshapi/grpc-rest-gateway/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type S struct {
	upgrader websocket.Upgrader
}

func main() {
	listener, err := net.Listen("tcp", ":40000")
	if err != nil {
		log.Fatalf("failed to bind: %s", err)
	}

	server := grpc.NewServer()
	echo.RegisterEchoServiceServer(server, &EchoService{})
	integrationapi.RegisterQueryParamsTestServer(server, &integration.QueryParamsTestServer{})
	integrationapi.RegisterPathParamsTestServer(server, &integration.PathParamsTestServer{})
	integrationapi.RegisterPatchRequestTestServer(server, &integration.PatchRequestTestServer{})
	integrationapi.RegisterStreamingTestServer(server, &integration.StreamingTestServer{})
	reflection.Register(server)

	connection, err := grpc.Dial(":40000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial: %s", err)
	}

	restGateway := gateway.NewServeMux()
	integrationapi.RegisterQueryParamsTestHandler(context.Background(), restGateway, connection)
	integrationapi.RegisterPathParamsTestHandler(context.Background(), restGateway, connection)
	integrationapi.RegisterPatchRequestTestHandler(context.Background(), restGateway, connection)
	integrationapi.RegisterStreamingTestHandler(context.Background(), restGateway, connection)

	go func() {
		log.Printf("starting HTTP on port 4000...")
		if err := http.ListenAndServe(":4000", restGateway); err != nil {
			log.Fatalf("failed to start HTTP Rest Gateway service: %s", err)
		}
	}()

	log.Printf("starting gRPC on port 40000...")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to start gRPC server: %s", err)
	}
}
