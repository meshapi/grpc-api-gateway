package grpctest

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type RegisterFunc func(*grpc.Server)

// Manager is a manager that controls a test gRPC server and the gateway that runs services.
//
// NOTE: This manager works with a *testing.T object to log failurs and mark tests as failed. This field is optional
// and when it is not available, panics are used instead. This is not a very idiomatic behavior however it simplifies
// the tests and these panics are indeed unexpected behavior that require author's attention as soon as possible.
type Manager struct {
	// RegisterServicesFunc will be called when the server is created to register services. Every new session calls this
	// function.
	RegisterServicesFunc RegisterFunc

	name       string
	server     *grpc.Server
	listener   *bufconn.Listener
	t          *testing.T
	connection *grpc.ClientConn
	lock       sync.Mutex
}

// NewManager creates a new manager instance without starting it.
func NewManager(t *testing.T, name string, registerFunc RegisterFunc) *Manager {
	return &Manager{
		t:                    t,
		name:                 name,
		RegisterServicesFunc: registerFunc,
	}
}

func (m *Manager) SetT(tt *testing.T) {
	m.t = tt
}

func (m *Manager) info(format string, args ...any) {
	if m.t != nil {
		m.t.Logf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (m *Manager) fatal(format string, args ...any) {
	if m.t != nil {
		m.t.Fatalf(format, args...)
	} else {
		panic(fmt.Sprintf(format, args...))
	}
}

// Active returns whether or not the server is running and available.
func (m *Manager) Active() bool {
	return m.server != nil
}

// Start is idempotent. If an existing server is available, it does nothing.
func (m *Manager) Start() {
	if m.Active() {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	// a secondary check is needed because it is possible that a server
	// instance is created while waiting for the lock.
	if m.Active() {
		return
	}

	m.server = grpc.NewServer()

	if m.RegisterServicesFunc == nil {
		m.info("RegisterServicesFunc is not set so no gRPC service will be registered")
	} else {
		m.RegisterServicesFunc(m.server)
	}

	m.listener = bufconn.Listen(10 * 1024 * 1024)
	go func() {
		defer m.listener.Close()

		if err := m.server.Serve(m.listener); err != nil {
			m.fatal("gRPC server %q failed: %s", m.name, err)
		}

		m.server = nil
		m.listener = nil
		m.connection = nil
	}()
}

func (m *Manager) Stop() {
	if m.Active() {
		m.server.Stop()
	}
}

// ClientConnection creates a gRPC client connection for the current server.
func (m *Manager) ClientConnection() *grpc.ClientConn {
	if m.connection != nil {
		return m.connection
	}

	if !m.Active() {
		m.fatal("gRPC server is not running, so no client can be created for it")
		return nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.connection != nil {
		return m.connection
	}

	dialer := func(context.Context, string) (net.Conn, error) {
		conn, err := m.listener.Dial()
		if err != nil {
			return conn, fmt.Errorf("failed to create buffer connection: %w", err)
		}

		return conn, nil
	}

	grpcOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	connection, err := grpc.Dial("bufnet", grpcOpts...)
	if err != nil {
		m.fatal("failed to create buffer connection: %s", err)
		return nil
	}

	m.connection = connection
	return connection
}
