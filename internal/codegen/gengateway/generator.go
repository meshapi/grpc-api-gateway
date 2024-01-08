package gengateway

import (
	"fmt"

	"github.com/meshapi/grpc-rest-gateway/internal/codegen/descriptor"
	"google.golang.org/protobuf/compiler/protogen"
)

// Position describes a specific line in a file.
type Position struct {
	// FileName is the name of the file.
	FileName string
}

type Failure struct {
	// Message holds the failure message.
	Message string

	// Position holds the position of the failure.
	Position Position
}

type GeneratedFile struct {
}

type Result struct {
	Files    []GeneratedFile
	Failures []Failure
}

type Generator struct {
	Options

	registry *descriptor.Registry
}

func New(options Options) *Generator {
	descriptorRegistry := descriptor.NewRegistry()
	descriptorRegistry.GatewayFileLoadOptions = options.GatewayFileLoadOptions
	descriptorRegistry.SearchPath = options.SearchPath

	return &Generator{Options: options, registry: descriptorRegistry}
}

func (g *Generator) LoadFromPlugin(gen *protogen.Plugin) error {
	if err := g.registry.LoadFromPlugin(gen); err != nil {
		return fmt.Errorf("failed to load proto files: %w", err)
	}

	return nil
}

func (g *Generator) Generate(targets []*descriptor.File) (Result, error) {
	return Result{}, nil
}
