package websocket

// Connection defines the server-side behavior of a streaming websocket connnection.
type Connection interface {
	// SendMessage sends a message. On error, SendMessage aborts the stream and the
	// error is returned directly.
	//
	// SendMessage blocks until:
	//   - There is sufficient flow control to schedule m with the transport, or
	//   - The stream is done, or
	//   - The stream breaks.
	//
	// SendMessage does not wait until the message is received by the client.
	// An untimely stream closure may result in lost messages.
	//
	// It is safe to have a goroutine calling SendMessage and another goroutine
	// calling ReceiveMessage on the same stream at the same time, but it is not safe
	// to call SendMessage on the same stream in different goroutines.
	//
	// It is not safe to modify the message after calling SendMessage. Tracing
	// libraries and stats handlers may use the message lazily.
	SendMessage(data []byte) error

	// ReceiveMessage blocks until it receives a message into m or the stream is
	// done. It returns io.EOF when the client has performed a CloseSend. On
	// any non-EOF error, the stream is aborted and the error contains the
	// RPC status.
	//
	// It is safe to have a goroutine calling SendMessage and another goroutine
	// calling ReceiveMessage on the same stream at the same time, but it is not
	// safe to call ReceiveMessage on the same stream in different goroutines.
	ReceiveMessage() ([]byte, error)

	// Close closes the connection immediately.
	//
	// NOTE: This method must be idempotent.
	Close() error
}
