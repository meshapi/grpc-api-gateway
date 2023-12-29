package descriptor

import "flag"

// GatewayFileLoadOptions holds the gateway config file loading options.
type GatewayFileLoadOptions struct {
	// GlobalGatewayConfigFile points to the global gateway config file.
	GlobalGatewayConfigFile string

	// FilePattern holds the file pattern for loading gateway config files.
	//
	// This pattern must not include the extension and the priority is yaml, yml and finally json.
	FilePattern string
}

// AddFlags adds command line flags to update this gateway loading options.
func (g *GatewayFileLoadOptions) AddFlags(flags *flag.FlagSet) {
	flags.StringVar(
		&g.GlobalGatewayConfigFile,
		"gateway-config",
		g.GlobalGatewayConfigFile,
		"(optional) path to the gateway config file that gets loaded first.")

	flags.StringVar(
		&g.FilePattern,
		"gateway-config-pattern",
		g.FilePattern,
		"gateway file pattern (without the extension segment) that gets used to try and load a gateway config file"+
			" for each proto file containing service definitions. yaml, yml and finally json file extensions will be tried.")
}

// DefaultGatewayLoadOptions holds the default gateway config loading options.
func DefaultGatewayLoadOptions() GatewayFileLoadOptions {
	return GatewayFileLoadOptions{
		GlobalGatewayConfigFile: "",
		FilePattern:             "{}_gateway",
	}
}
