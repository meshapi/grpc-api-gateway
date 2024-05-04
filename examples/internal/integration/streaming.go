package integration

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/meshapi/grpc-api-gateway/examples/internal/gen/integration"
)

type StreamingTestServer struct {
	integration.UnimplementedStreamingTestServer

	users         map[string]integration.StreamingTest_ChatServer
	broadcastLock sync.Mutex
}

func NewStreamingTestServer() *StreamingTestServer {
	return &StreamingTestServer{
		users: map[string]integration.StreamingTest_ChatServer{},
	}
}

func (s *StreamingTestServer) Add(server integration.StreamingTest_AddServer) error {
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

func (s *StreamingTestServer) Generate(req *integration.GenerateRequest, server integration.StreamingTest_GenerateServer) error {
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

		time.Sleep(time.Duration(int64(req.Wait) * int64(time.Second)))
	}

	return nil
}

func (s *StreamingTestServer) BulkCapitalize(server integration.StreamingTest_BulkCapitalizeServer) error {
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

func (s *StreamingTestServer) NotifyMe(
	req *integration.NotifyMeRequest, server integration.StreamingTest_NotifyMeServer) error {

	index := 0
	for {
		err := server.SendMsg(&integration.NotifyMeResponse{
			Text: fmt.Sprintf("Notification #%d", index+1),
		})
		if err != nil {
			return fmt.Errorf("failed to send: %w", err)
		}

		time.Sleep(time.Duration(int64(req.Wait) * int64(time.Second)))
		index++
	}
}

func (s *StreamingTestServer) Chat(server integration.StreamingTest_ChatServer) error {
	userName := fmt.Sprintf("user-%d", rand.Intn(10000))
	s.users[userName] = server
	defer func() {
		delete(s.users, userName)
		s.broadcast("dispatcher", fmt.Sprintf("%q left", userName))
	}()
	s.broadcast("dispatcher", fmt.Sprintf("user %q joined", userName))
	for {
		req, err := server.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		s.broadcast(userName, req.Text)
	}
}

func (s *StreamingTestServer) broadcast(source, text string) {
	s.broadcastLock.Lock()
	defer s.broadcastLock.Unlock()
	for userName, stream := range s.users {
		if err := stream.Send(&integration.ChatResponse{Text: text, Source: source}); err != nil {
			log.Printf("failed to send message to %q: %s", userName, err)
		}
	}
}
