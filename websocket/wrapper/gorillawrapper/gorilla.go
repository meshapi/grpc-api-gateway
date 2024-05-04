package gorillawrapper

import (
	"io"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Connection is a wrapper around a gorilla connection that conforms to websocket.Connection interface and can be used
// in gateways.
type Connection struct {
	underlyingConnection *websocket.Conn

	// CloseTimeout is the amount of time to wait after sending a closure status.
	CloseTimeout time.Duration
}

// New creates a new Connection instance with a valid gorilla connection.
func New(conn *websocket.Conn) Connection {
	return Connection{
		underlyingConnection: conn,
		CloseTimeout:         30 * time.Second,
	}
}

func (c Connection) SendMessage(data []byte) error {
	return c.underlyingConnection.WriteMessage(websocket.TextMessage, data)
}

func (c Connection) SendClose() error {
	return c.underlyingConnection.WriteControl(websocket.CloseMessage, nil, time.Now().Add(c.CloseTimeout))
}

func (c Connection) ReceiveMessage() ([]byte, error) {
	_, data, err := c.underlyingConnection.ReadMessage()
	if err != nil && isClosedConnectionError(err) {
		return nil, io.EOF
	}

	return data, err
}

func (c Connection) Close() error {
	err := c.underlyingConnection.Close()
	if err != nil && isClosedConnectionError(err) {
		return nil
	}
	return err
}

func isClosedConnectionError(err error) bool {
	if websocket.IsCloseError(err,
		websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway) {
		return true
	}
	return strings.Contains(err.Error(), "use of closed network connection")
}
