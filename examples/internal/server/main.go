package main

import (
	"log"
	"net"
	"net/http"

	"github.com/meshapi/grpc-rest-gateway/examples/internal/proto/echo"
	"github.com/meshapi/grpc-rest-gateway/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	listener, err := net.Listen("tcp", ":40000")
	if err != nil {
		log.Fatalf("failed to bind: %s", err)
	}

	server := grpc.NewServer()
	echo.RegisterEchoServiceServer(server, &EchoService{})
	reflection.Register(server)

	//connection, err := grpc.Dial(":40000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	//if err != nil {
	//  log.Fatalf("failed to dial: %s", err)
	//}

	restGateway := gateway.NewServeMux()

	go func() {
		log.Printf("starting HTTP on port 4000...")
		if err := http.ListenAndServe(":4000", restGateway); err != nil {
			log.Printf("failed to start HTTP Rest Gateway service: %s", err)
		}
	}()

	log.Printf("starting gRPC on port 40000...")
	server.Serve(listener)
}
