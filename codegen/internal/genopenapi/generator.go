package genopenapi

import "github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"

type Generator struct {
	Options

	registry *descriptor.Registry

	// httpEndpointsMap is used to find duplicate HTTP specifications.
	//httpEndpointsMap map[endpointAnnotation]struct{}
}

func New(descriptorRegistry *descriptor.Registry, options Options) *Generator {
	return &Generator{
		Options:  options,
		registry: descriptorRegistry,
	}
}

func (g *Generator) Generate(v any) ([]descriptor.ResponseFile, error) {
	return nil, nil
}
