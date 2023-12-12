package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/meshapi/grpc-rest-gateway/internal/codegen/plugin"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type pingService struct {
	plugin.UnimplementedRestGatewayPluginServer
}

func (p pingService) Ping(ctx context.Context, req *plugin.PingRequest) (*plugin.PingResponse, error) {
	return &plugin.PingResponse{Text: req.Text}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":30000")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	defer listener.Close()

	go func() {
		pluginInfo := &plugin.PluginInfo{
			Connection: &plugin.PluginInfo_Tcp{
				Tcp: &plugin.TCPConnection{Address: ":30000"}},
			RegisteredCallbacks: []string{"feature-a", "feature-b"},
		}
		data, err := proto.Marshal(pluginInfo)
		if err != nil {
			log.Fatalf("failed to marshal: %s", err)
		}
		fmt.Println(len(data))
		if _, err := os.Stdout.Write(data); err != nil {
			log.Fatalf("failed to marshal: %s", err)
		}

	}()

	server := grpc.NewServer()
	plugin.RegisterRestGatewayPluginServer(server, &pingService{})
	server.Serve(listener)
}
