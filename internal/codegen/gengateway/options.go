package gengateway

import (
	"fmt"
	"strings"
)

type PathParameterSeparator uint8

func (p PathParameterSeparator) String() string {
	switch p {
	case PathParameterSeparatorCSV:
		return "csv"
	case PathParameterSeparatorTSV:
		return "tsv"
	case PathParameterSeparatorSSV:
		return "ssv"
	case PathParameterSeparatorPipes:
		return "pipes"
	default:
		return "<unknown>"
	}
}

func (p *PathParameterSeparator) Set(value string) error {
	switch strings.ToLower(value) {
	case "csv":
		*p = PathParameterSeparatorCSV
	case "tsv":
		*p = PathParameterSeparatorTSV
	case "ssv":
		*p = PathParameterSeparatorSSV
	case "pipes":
		*p = PathParameterSeparatorPipes
	default:
		return fmt.Errorf("unrecognized value: '%s'. Allowed values are 'cav', 'pipes', 'ssv' and 'tsv'.", value)
	}

	return nil
}

const (
	PathParameterSeparatorCSV = iota
	PathParameterSeparatorPipes
	PathParameterSeparatorSSV
	PathParameterSeparatorTSV
)

// Options are the options for the code generator.
type Options struct {
	// RegisterFunctionSuffix is used to construct names of the generated Register*<Suffix> methods.
	RegisterFunctionSuffix string

	// UseHTTPRequestContext controls whether or not HTTP request's context gets used.
	UseHTTPRequestContext bool

	// AllowDeleteBody indicates whether or not DELETE methods can have bodies.
	AllowDeleteBody bool

	// RepeatedPathParameterSeparator determines how repeated fields should be split when used in path segments.
	RepeatedPathParameterSeparator PathParameterSeparator

	// AllowPatchFeature determines whether to use PATCH feature involving update masks
	// (using using google.protobuf.FieldMask).
	AllowPatchFeature bool

	// OmitPackageDoc indicates whether or not package commments should be included in generated code.
	OmitPackageDoc bool

	// Standalone generates a standalone gateway package, which imports the target service package.
	Standalone bool
}

// DefaultOptions returns the default options.
func DefaultOptions() Options {
	return Options{
		RegisterFunctionSuffix:         "Handler",
		UseHTTPRequestContext:          true,
		AllowDeleteBody:                false,
		RepeatedPathParameterSeparator: PathParameterSeparatorCSV,
		AllowPatchFeature:              true,
		OmitPackageDoc:                 false,
		Standalone:                     false,
	}
}
