package descriptor

import (
	"flag"
)

// RegistryOptions holds all the options for the descriptor registry.
type RegistryOptions struct {
	// GatewayFileLoadOptions holds gateway config file loading options.
	GatewayFileLoadOptions GatewayFileLoadOptions

	// SearchPath is the directory that is used to look for gateway configuration files.
	//
	// this search path can be relative or absolute, if relative, it will be from the current working directory.
	SearchPath string

	// WarnOnUnboundMethods emits a warning message if an RPC method has no mapping.
	WarnOnUnboundMethods bool

	// GenerateUnboundMethods controls whether or not unannotated RPC methods should be created as part of the proxy.
	GenerateUnboundMethods bool

	// AllowDeleteBody indicates whether or not DELETE methods can have bodies.
	AllowDeleteBody bool

	// Standalone mode is to prepare for generation of Go files as a standalone package.
	Standalone bool
}

// GatewayFileLoadOptions holds the gateway config file loading options.
type GatewayFileLoadOptions struct {
	// GlobalGatewayConfigFile points to the global gateway config file.
	GlobalGatewayConfigFile string

	// FilePattern holds the file pattern for loading gateway config files.
	//
	// This pattern must not include the extension and the priority is yaml, yml and finally json.
	FilePattern string
}

// addFlags adds command line flags to update this gateway loading options.
func (g *GatewayFileLoadOptions) addFlags(flags *flag.FlagSet) {
	flags.StringVar(
		&g.GlobalGatewayConfigFile,
		"gateway_config",
		g.GlobalGatewayConfigFile,
		"(optional) path to the gateway config file that gets loaded first.")

	flags.StringVar(
		&g.FilePattern,
		"gateway_config_pattern",
		g.FilePattern,
		"gateway file pattern (without the extension segment) that gets used to try and load a gateway config file"+
			" for each proto file containing service definitions. yaml, yml and finally json file extensions will be tried.")
}

// defaultGatewayLoadOptions holds the default gateway config loading options.
func defaultGatewayLoadOptions() GatewayFileLoadOptions {
	return GatewayFileLoadOptions{
		GlobalGatewayConfigFile: "",
		FilePattern:             "{{ .Path }}_gateway",
	}
}

func DefaultRegistryOptions() RegistryOptions {
	return RegistryOptions{
		GatewayFileLoadOptions: defaultGatewayLoadOptions(),
		SearchPath:             ".",
		AllowDeleteBody:        false,
	}
}

// AddFlags adds command line flags to update this gateway loading options.
func (r *RegistryOptions) AddFlags(flags *flag.FlagSet) {
	r.GatewayFileLoadOptions.addFlags(flags)

	flags.StringVar(
		&r.SearchPath,
		"config_search_path",
		r.SearchPath,
		"gateway config search path is the directory (relative or absolute) from the current working directory that contains"+
			" the gateway config files.")

	flags.BoolVar(
		&r.WarnOnUnboundMethods, "warn_on_unbound_methods", r.WarnOnUnboundMethods,
		"emits a warning message if an RPC method has no mapping.")

	flags.BoolVar(
		&r.GenerateUnboundMethods, "generate_unbound_methods", r.GenerateUnboundMethods,
		"controls whether or not unannotated RPC methods should be created as part of the proxy.")

	flag.BoolVar(
		&r.AllowDeleteBody, "allow_delete_body", r.AllowDeleteBody,
		"unless set, HTTP DELETE methods may not have a body")
}
