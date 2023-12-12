package main

import (
	"context"
	"log"

	"github.com/meshapi/grpc-rest-gateway/internal/codegen/plugin"
)

func main() {
	manager := plugin.NewManager("./plugin", &plugin.GeneratorInfo{
		Version: &plugin.Version{
			Major: 0,
			Minor: 1,
			Patch: 0,
		},
		Generator:         plugin.Generator_Generator_RestGateway,
		SupportedFeatures: []string{"feature-a", "feature-b"},
	})

	client, err := manager.InitRestGateway(context.Background())
	if err != nil {
		log.Fatalf("failed to initialize plugin: %s", err)
	}

	resp, err := client.Ping(context.Background(), &plugin.PingRequest{Text: "Hi"})
	if err != nil {
		log.Fatalf("plugin failed to respond to ping: %s", err)
	}

	log.Printf("RESPONSE: %s", resp.Text)
}
