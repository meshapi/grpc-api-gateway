package gengateway

import (
	"fmt"
	"path"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"golang.org/x/tools/imports"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

type Generator struct {
	Options

	registry    *descriptor.Registry
	baseImports []descriptor.GoPackage

	// httpEndpointsMap is used to find duplicate HTTP specifications.
	httpEndpointsMap map[endpointAnnotation]struct{}
}

type endpointAnnotation struct {
	Method       string
	PathTemplate string
	Service      *descriptor.Service
}

func New(descriptorRegistry *descriptor.Registry, options Options) *Generator {
	var imports []descriptor.GoPackage
	for _, pkgpath := range []string{
		"context",
		"io",
		"net/http",
		"github.com/meshapi/grpc-rest-gateway/gateway",
		"github.com/meshapi/grpc-rest-gateway/iofactory",
		"github.com/meshapi/grpc-rest-gateway/partialfieldmask",
		"github.com/meshapi/grpc-rest-gateway/protoconvert",
		"github.com/meshapi/grpc-rest-gateway/protopath",
		"github.com/meshapi/grpc-rest-gateway/trie",
		"google.golang.org/protobuf/proto",
		"google.golang.org/grpc",
		"google.golang.org/grpc/codes",
		"google.golang.org/grpc/grpclog",
		"google.golang.org/grpc/metadata",
		"google.golang.org/grpc/status",
		"github.com/julienschmidt/httprouter",
	} {
		pkg := descriptor.GoPackage{
			Path: pkgpath,
			Name: path.Base(pkgpath),
		}
		if ok := descriptorRegistry.ReserveGoPackageAlias(pkg.Name, pkg.Path); !ok {
			for i := 0; ; i++ {
				alias := fmt.Sprintf("%s_%d", pkg.Name, i)
				if ok := descriptorRegistry.ReserveGoPackageAlias(alias, pkg.Path); !ok {
					continue
				}
				pkg.Alias = alias
				break
			}
		}
		imports = append(imports, pkg)
	}

	return &Generator{
		Options:          options,
		baseImports:      imports,
		registry:         descriptorRegistry,
		httpEndpointsMap: make(map[endpointAnnotation]struct{}),
	}
}

func (g *Generator) Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	var files []*descriptor.ResponseFile
	for _, file := range targets {
		code, err := g.generate(file)
		if err != nil {
			return nil, fmt.Errorf("error generating rest gateway for %q: %w", file.GetName(), err)
		}
		if code == "" { // if there is no code for this target, move on.
			continue
		}

		formatted, err := imports.Process(file.GetName(), []byte(code), nil)
		if err != nil {
			grpclog.Errorf("%v: %s", err, code)
			return nil, err
		}

		files = append(files, &descriptor.ResponseFile{
			GoPkg: file.GoPkg,
			CodeGeneratorResponse_File: &pluginpb.CodeGeneratorResponse_File{
				Name:    proto.String(file.GeneratedFilenamePrefix + ".pb.rgw.go"),
				Content: proto.String(string(formatted)),
			},
		})
	}

	return files, nil
}

func (g *Generator) CheckDuplicateEndpoint(
	method, pathTemplate string, service *descriptor.Service) error {

	annotation := endpointAnnotation{
		Method:       method,
		PathTemplate: pathTemplate,
		Service:      service,
	}

	if _, ok := g.httpEndpointsMap[annotation]; ok {
		return fmt.Errorf("duplicate annotation: method=%s, template=%s", method, pathTemplate)
	}
	g.httpEndpointsMap[annotation] = struct{}{}
	return nil
}

func (g *Generator) generate(file *descriptor.File) (string, error) {
	pkgSeen := make(map[string]bool)
	var imports []descriptor.GoPackage
	for _, pkg := range g.baseImports {
		pkgSeen[pkg.Path] = true
		imports = append(imports, pkg)
	}

	for _, svc := range file.Services {
		for _, m := range svc.Methods {
			imports = append(imports, g.addEnumPathParamImports(file, m, pkgSeen)...)
			pkg := m.RequestType.File.GoPkg
			if len(m.Bindings) == 0 ||
				pkg == file.GoPkg || pkgSeen[pkg.Path] {
				continue
			}
			pkgSeen[pkg.Path] = true
			imports = append(imports, pkg)
		}
	}
	params := param{
		File:               file,
		Imports:            imports,
		UseRequestContext:  g.Options.UseHTTPRequestContext,
		RegisterFuncSuffix: g.Options.RegisterFunctionSuffix,
		AllowPatchFeature:  g.Options.AllowPatchFeature,
	}
	if g.registry != nil {
		params.OmitPackageDoc = g.Options.OmitPackageDoc
	}

	return g.applyTemplate(params, g.registry)
}

// addEnumPathParamImports handles adding import of enum path parameter go packages
func (g *Generator) addEnumPathParamImports(file *descriptor.File, m *descriptor.Method, pkgSeen map[string]bool) []descriptor.GoPackage {
	var imports []descriptor.GoPackage
	for _, b := range m.Bindings {
		for _, p := range b.PathParameters {
			e, err := g.registry.LookupEnum("", p.Target.GetTypeName())
			if err != nil {
				continue
			}

			pkg := e.File.GoPkg
			if pkg == file.GoPkg || pkgSeen[pkg.Path] {
				continue
			}

			pkgSeen[pkg.Path] = true
			imports = append(imports, pkg)
		}
	}

	return imports
}
