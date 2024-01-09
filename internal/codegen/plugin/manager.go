package plugin

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/meshapi/grpc-rest-gateway/api/codegen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/proto"
)

const DefaultTimeout = 5 * time.Second

const (
	envVersion       = "GRPC_REST_GATEWAY_GENERATOR_VERSION"
	envGeneratorType = "GRPC_REST_GATEWAY_GENERATOR_TYPE"
)

type initConfig struct {
	Timeout time.Duration
}

type Option interface {
	apply(*initConfig)
}

type optionFunc func(*initConfig)

func (o optionFunc) apply(c *initConfig) {
	o(c)
}

// WithTimeout sets a plugin activation timeout.
func WithTimeout(duration time.Duration) Option {
	return optionFunc(func(c *initConfig) {
		c.Timeout = duration
	})
}

func getGeneratorType(generator codegen.Generator) string {
	switch generator {
	case codegen.Generator_Generator_RestGateway:
		return "REST_GATEWAY"
	case codegen.Generator_Generator_OpenAPI:
		return "OPEN_API"
	default:
		panic("Using unknown generator is not acceptable")
	}
}

type Callbacks []string

func (c Callbacks) Has(callback string) bool {
	for _, item := range c {
		if item == callback {
			return true
		}
	}

	return false
}

type Client struct {
	Gateway             codegen.RestGatewayPluginClient
	RegisteredCallbacks Callbacks
}

type Manager struct {
	initConfig
	info *codegen.GeneratorInfo
	path string

	connection *grpc.ClientConn
	pluginInfo *codegen.PluginInfo
	cmd        *exec.Cmd
	lock       sync.Mutex
}

func NewManager(path string, info *codegen.GeneratorInfo, opts ...Option) *Manager {
	config := initConfig{Timeout: DefaultTimeout}

	for _, opt := range opts {
		opt.apply(&config)
	}

	return &Manager{initConfig: config, path: path, info: info}
}

// InitConnection initializes a plugin connection by starting a subprocess at path and exchanging connection info.
func (m *Manager) InitConnection(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// if already active, do not initialize a new connection.
	if m.connection != nil {
		return nil
	}

	pluginPath, err := exec.LookPath(m.path)
	if err != nil {
		return fmt.Errorf("failed to lookup path '%s': %w", m.path, err)
	}

	m.cmd = exec.Command(pluginPath)

	// add environment variables.
	m.cmd.Env = append(m.cmd.Env,
		fmt.Sprintf("%s=%d.%d.%d", envVersion, m.info.Version.Major, m.info.Version.Minor, m.info.Version.Patch))
	m.cmd.Env = append(m.cmd.Env, fmt.Sprintf("%s=%s", envGeneratorType, getGeneratorType(m.info.Generator)))

	writer, err := m.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to connect to plugin process's stdin: %w", err)
	}

	reader, err := m.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to connect to plugin process's stdout: %w", err)
	}

	stderr := &bytes.Buffer{}
	m.cmd.Stderr = stderr

	// call the process.
	if err := m.cmd.Start(); err != nil {
		m.cmd = nil
		return fmt.Errorf("failed to start plugin process: %w", err)
	}
	processExitLogChan := make(chan struct{})

	go waitAndLogProcess(m.cmd, stderr, processExitLogChan)

	if err := writeGeneratorInfoAndClose(writer, m.info); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, m.Timeout)
	defer cancel()

	errChan := make(chan error)
	go func() {
		info, err := readPluginInfo(reader)
		if err != nil {
			errChan <- fmt.Errorf("failed to read plugin info from stdout: %w", err)
			return
		}

		conn, err := dialPluginConnection(ctx, info)
		if err != nil {
			errChan <- fmt.Errorf("failed to dial and prepare grpc connection to plugin: %w", err)
			return
		}

		m.pluginInfo = info
		m.connection = conn
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		if err != nil {
			// wait for the process to exit but no more than a few moments.
			select {
			case <-time.After(1 * time.Second):
			case <-processExitLogChan:
			}
		}
		return err
	case <-time.After(m.Timeout): // timeout has happened so drop everything.
		tryKillingProcess(m.cmd)
		return errors.New("plugin initialization timed out")
	case <-ctx.Done(): // context is canceled so drop everything.
		tryKillingProcess(m.cmd)
		return errors.New("context is canceled")
	}
}

func (m *Manager) InitClient(ctx context.Context) (Client, error) {
	err := m.InitConnection(ctx)
	if err != nil {
		return Client{}, err
	}

	return Client{
		Gateway:             codegen.NewRestGatewayPluginClient(m.connection),
		RegisteredCallbacks: m.pluginInfo.RegisteredCallbacks,
	}, nil
}

// Kill would attempt to kill the current plugin process if there is one available.
func (m *Manager) Kill() {
	tryKillingProcess(m.cmd)
}

func tryKillingProcess(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		if err := cmd.Process.Kill(); err != nil {
			grpclog.Warningf("failed to kill plugin '%s' process: %s", cmd.Path, err)
		}
	}
}

func dialPluginConnection(ctx context.Context, info *codegen.PluginInfo) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	target := ""

	switch connectionInfo := info.Connection.(type) {
	case *codegen.PluginInfo_Tcp:
		target = connectionInfo.Tcp.Address
	case *codegen.PluginInfo_UnixSocket:
		target = connectionInfo.UnixSocket.Socket
		if !strings.HasPrefix(target, "unix://") {
			target = "unix://" + target
		}
	default:
		return nil, errors.New("unsupported connection method, accepting only TCP/Unix")
	}

	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial plugin service at %s: %w", target, err)
	}

	return conn, nil
}

func readPluginInfo(reader io.Reader) (*codegen.PluginInfo, error) {
	var length uint64
	if _, err := fmt.Fscanf(reader, "%d", &length); err != nil {
		return nil, fmt.Errorf("failed to capture message length: %w", err)
	}

	if length > 1e6 {
		return nil, fmt.Errorf("plugin initialization message is too large (>1mb): %d", length)
	}

	buffer := make([]byte, length)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return nil, fmt.Errorf("failed to read buffer: %w", err)
	}

	info := &codegen.PluginInfo{}
	if err := proto.Unmarshal(buffer, info); err != nil {
		return nil, fmt.Errorf("failed to marshal PluginInfo: %w", err)
	}

	return info, nil
}

func writeGeneratorInfoAndClose(writer io.WriteCloser, info *codegen.GeneratorInfo) error {
	defer writer.Close()

	generatorInfo, err := proto.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal generator info: %w", err)
	}

	if _, err := writer.Write(generatorInfo); err != nil {
		return fmt.Errorf("failed to write to the process stdin: %w", err)
	}

	return nil
}

func waitAndLogProcess(cmd *exec.Cmd, stderr io.Reader, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
	}()

	err := cmd.Wait()
	if err == nil {
		return
	}

	exitError, ok := err.(*exec.ExitError)
	if !ok || exitError.ExitCode() == 0 {
		return
	}

	data, err := io.ReadAll(stderr)
	if err != nil {
		grpclog.Warningf("failed to read stderr: %s", err)
	}

	errorOutput := ""
	if data != nil {
		outputWriter := &strings.Builder{}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			fmt.Fprintf(outputWriter, "[%s] %s\n", cmd.Path, scanner.Text())
		}
		errorOutput = outputWriter.String()
	} else {
		errorOutput = "<no stderr output>"
	}

	grpclog.Errorf("plugin '%s' failed with exit code %d:\n%s", cmd.Path, exitError.ExitCode(), errorOutput)
}
