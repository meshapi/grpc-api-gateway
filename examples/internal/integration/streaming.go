package integration

import (
	"fmt"
	"io"
	"strings"

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

		result.Sum += req.Value
		result.Count++
	}

	if err := server.SendAndClose(result); err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}

	return nil
}

func (s StreamingTestServer) Generate(req *integration.GenerateRequest, server integration.StreamingTest_GenerateServer) error {
	var current int32

	var err error
	for i := 0; i < int(req.Count); i++ {
		current = current*2 + 1

		err = server.Send(&integration.GenerateResponse{
			Index: int32(i),
			Value: current,
		})

		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	return nil
}

func (s StreamingTestServer) BulkCapitalize(server integration.StreamingTest_BulkCapitalizeServer) error {
	count := 0
	for {
		req, err := server.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive: %w", err)
		}

		count++
		err = server.Send(&integration.BulkCapitalizeResponse{
			Index: int32(count),
			Text:  strings.ToUpper(req.Text),
		})
		if err != nil {
			return fmt.Errorf("failed to send: %w", err)
		}
	}

	return nil
}
