package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/echo"
	integrationapi "github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
	"github.com/meshapi/grpc-rest-gateway/examples/internal/integration"
	"github.com/meshapi/grpc-rest-gateway/gateway"
	ws "github.com/meshapi/grpc-rest-gateway/websocket"
	"github.com/meshapi/grpc-rest-gateway/websocket/backends/gorillawrapper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"

	"github.com/gorilla/websocket"
)

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
	integrationapi.RegisterStreamingTestServer(server, integration.NewStreamingTestServer())
	reflection.Register(server)

	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout))

	connection, err := grpc.Dial(":40000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial: %s", err)
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	websocketUpgradeFunc := gateway.WebsocketUpgradeFunc(func(w http.ResponseWriter, r *http.Request) (ws.Connection, error) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws error: %s", err)
			return nil, fmt.Errorf("failed to upgrade: %w", err)
		}

		return gorillawrapper.New(c), nil
	})

	restGateway := gateway.NewServeMux(gateway.WithWebsocketUpgrader(websocketUpgradeFunc))
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
