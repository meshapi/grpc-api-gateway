package integration

import (
	"fmt"
	"io"
	"log"

	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
)

type StreamingTestServer struct {
	integration.UnimplementedStreamingTestServer
}

func (s StreamingTestServer) Add(server integration.StreamingTest_AddServer) error {

	result := &integration.AddResponse{}

	for {
		req, err := server.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read client streaming request: %w", err)
		}
		log.Printf("received: %+v", req)

		result.Sum += req.Value
		result.Count++
	}

	if err := server.SendAndClose(result); err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}

	return nil
}
