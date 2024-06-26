package genopenapi

import (
	"github.com/meshapi/grpc-api-gateway/codegen/internal/descriptor"
)

// Options are the options for the code generator.
type Options struct {
	// RepeatedPathParameterSeparator determines how repeated fields should be split when used in path segments.
	RepeatedPathParameterSeparator descriptor.PathParameterSeparator

	// AllowPatchFeature determines whether to use PATCH feature involving update masks
	// (using using google.protobuf.FieldMask).
	AllowPatchFeature bool

	// IncludeServicesOnly generates OpenAPI output only for bound service endpoints and will omit all unused models.
	IncludeServicesOnly bool

	// OutputMode indicates the mode of OpenAPI output generation, to merge all definitions into one file, per service or
	// per proto file.
	OutputMode OutputMode

	// OperationIDMode determines the mode of operation IDs that get generated.
	OperationIDMode OperationIDMode

	// OutputFileName is the OpenAPI output file name after merging all files.
	// Only applicable when output mode is "merge".
	OutputFileName string

	// OutputFormat is the resulting OpenAPI format.
	OutputFormat OutputFormat

	// FieldNameMode determines what naming convention the fields in the OpenAPI schemas get.
	FieldNameMode FieldNameMode

	// IncludePackageInTags includes the fully qualified service name (FQSN) in the tags of each operation.
	IncludePackageInTags bool

	// DisableServiceTags disables generation of service tags in OpenAPI, useful to avoid exposing gRPC services.
	DisableServiceTags bool

	// SchemaNamingStrategy holds the naming strategy for schema names in generated OpenAPI output.
	SchemaNamingStrategy SchemaNamingStrategy

	// UseGoTemplate allows using templates for summary, description, tags and links.
	//
	// TODO: include a link to the context avaialble for the evaluation.
	UseGoTemplate bool

	// GoTemplateArgs are additional template args that can be set. GoTemplate must be enabled in order to utilize this.
	GoTemplateArgs TemplateArgs

	// If set to true, proto doc strings get ignored.
	IgnoreComments bool

	// If set to true, all comment lines that start with (-- and end with --) get excluded.
	RemoveInternalComments bool

	// If set to true, the default error response does not get added to the responses.
	DisableDefaultErrors bool

	// If set to true, the default 200 successful response does not get added to the responses.
	DisableDefaultResponses bool

	// UseEnumNumbers uses numerical value of enums instead of strings.
	UseEnumNumbers bool

	// GlobalOpenAPIConfigFile points to the file that can be used to define global OpenAPI config file.
	GlobalOpenAPIConfigFile string

	// ConfigSearchPath holds the search path to use for looking up OpenAPI configs.
	ConfigSearchPath string

	// LocalPackageMode limits the config files to only targetting objects from their own proto package.
	LocalPackageMode bool

	// OpenAPIConfigFilePattern holds the file pattern for loading OpenAPI config files.
	//
	// This pattern must not include the extension and the priority is yaml, yml and finally json.
	OpenAPIConfigFilePattern string

	// OpenAPISeedFile holds an OpenAPI file in YAML/JSON format that will be used as a seed that will be merged
	// with the generated OpenAPI files.
	OpenAPISeedFile string

	// OmitEnumDefaultValue omits the default/unknown enum value.
	OmitEnumDefaultValue bool

	// VisibilitySelectors are a list of visibility selectors.
	VisibilitySelectors SelectorMap

	// MergeWithOverwrite will overwrite lists instead of appending.
	MergeWithOverwrite bool

	// OmitEmptyFiles avoids writing OpenAPI document files if the file does not contain at least one model or path.
	OmitEmptyFiles bool

	// FieldNullableMode configures the generation of nullable fields in the OpenAPI schemas.
	FieldNullableMode FieldNullableMode

	// FieldRequiredMode configures the generation of required field in the OpenAPI schemas.
	FieldRequiredMode FieldRequiredMode

	// WarnOnBrokenSelectors writes a warning instead of reporting errors when selectors to unmatching proto types is
	// used in config files.
	WarnOnBrokenSelectors bool
}

// DefaultOptions returns the default options.
func DefaultOptions() Options {
	return Options{
		RepeatedPathParameterSeparator: descriptor.PathParameterSeparatorCSV,
		AllowPatchFeature:              true,
		IncludeServicesOnly:            false,
		OutputMode:                     OutputModePerProtoFile,
		OutputFileName:                 "apidocs",
		OutputFormat:                   OutputFormatJSON,
		FieldNameMode:                  FieldNameModeJSON,
		IncludePackageInTags:           false,
		DisableServiceTags:             false,
		SchemaNamingStrategy:           SchemaNamingStrategySimple,
		UseGoTemplate:                  false,
		GoTemplateArgs:                 nil,
		IgnoreComments:                 false,
		RemoveInternalComments:         false,
		DisableDefaultErrors:           false,
		DisableDefaultResponses:        false,
		UseEnumNumbers:                 false,
		GlobalOpenAPIConfigFile:        "",
		ConfigSearchPath:               ".",
		OpenAPIConfigFilePattern:       "{{ .Path }}_gateway",
		OpenAPISeedFile:                "",
		OmitEnumDefaultValue:           false,
		VisibilitySelectors:            nil,
		MergeWithOverwrite:             true,
		OperationIDMode:                OperationIDModeServiceAndMethod,
		OmitEmptyFiles:                 false,
		FieldNullableMode:              FieldNullableModeOptionalLabel,
		FieldRequiredMode:              FieldRequiredModeDisabled,
		LocalPackageMode:               false,
	}
}
