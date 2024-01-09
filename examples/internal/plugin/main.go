package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/meshapi/grpc-rest-gateway/api/codegen"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type gatewayPlugin struct {
	codegen.UnimplementedRestGatewayPluginServer
}

func main() {
	listener, err := net.Listen("tcp", ":30000")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	defer listener.Close()

	go func() {
		pluginInfo := &codegen.PluginInfo{
			Connection: &codegen.PluginInfo_Tcp{
				Tcp: &codegen.TCPConnection{Address: ":30000"}},
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
	codegen.RegisterRestGatewayPluginServer(server, &gatewayPlugin{})
	_ = server.Serve(listener)
}
