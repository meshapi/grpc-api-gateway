package gengateway

import "github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"

// Options are the options for the code generator.
type Options struct {
	// RegisterFunctionSuffix is used to construct names of the generated Register*<Suffix> methods.
	RegisterFunctionSuffix string

	// UseHTTPRequestContext controls whether or not HTTP request's context gets used.
	UseHTTPRequestContext bool

	// RepeatedPathParameterSeparator determines how repeated fields should be split when used in path segments.
	RepeatedPathParameterSeparator descriptor.PathParameterSeparator

	// AllowPatchFeature determines whether to use PATCH feature involving update masks
	// (using using google.protobuf.FieldMask).
	AllowPatchFeature bool

	// OmitPackageDoc indicates whether or not package commments should be included in generated code.
	OmitPackageDoc bool

	// Standalone generates a standalone gateway package, which imports the target service package.
	Standalone bool

	// GenerateLocal generates code to work a server implementation directly.
	GenerateLocal bool
}

// DefaultOptions returns the default options.
func DefaultOptions() Options {
	return Options{
		RegisterFunctionSuffix:         "Handler",
		UseHTTPRequestContext:          true,
		RepeatedPathParameterSeparator: descriptor.PathParameterSeparatorCSV,
		AllowPatchFeature:              true,
		OmitPackageDoc:                 false,
		Standalone:                     false,
		GenerateLocal:                  false,
	}
}
