package main

import (
	"flag"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi"
)

// prepareOptions prepares a gen gateway options and adds necessary flags.
func prepareOptions() *genopenapi.Options {
	generatorOptions := genopenapi.DefaultOptions()

	flag.Var(
		&generatorOptions.RepeatedPathParameterSeparator, "repeated_path_param_separator",
		"configures how repeated fields should be split. Allowed values are 'csv', 'pipes', 'ssv', and 'tsv'.")

	flag.BoolVar(
		&generatorOptions.AllowPatchFeature, "allow_patch_feature", generatorOptions.AllowPatchFeature,
		"determines whether to use the PATCH feature involving update masks (using google.protobuf.FieldMask).")

	flag.BoolVar(
		&generatorOptions.IncludeServicesOnly, "include_services_only", generatorOptions.IncludeServicesOnly,
		"if true, only bound service methods and models needed to support them get generated.")

	flag.Var(
		&generatorOptions.OutputMode, "output_mode",
		"use 'merge' to merge all definitions into one file, 'proto' to generate one file per proto file,"+
			" 'service' to generate a separate document per gRPC service.")

	flag.Var(
		&generatorOptions.OperationIDMode, "operation_id_mode",
		"controls the mode of operation ids in the OpenAPI document."+
			" use 'service+method' for '<Service>_<Method>', 'method' for just the method name "+
			"and 'fqn' for the fully qualified name.")

	flag.StringVar(
		&generatorOptions.OutputFileName, "output_filename", generatorOptions.OutputFileName,
		"only applicable when using output mode 'merge'. It sets the file name of the generated OpenAPI document.")

	flag.Var(
		&generatorOptions.OutputFormat, "output_format",
		"controls the output format. Allowed values are 'yaml' and 'json'.")

	flag.Var(
		&generatorOptions.FieldNameMode, "field_name_mode",
		"controls the naming of fields in the OpenAPI schemas. Allowed values are 'proto' to "+
			"use the proto field names and 'json' to use the camel case JSON names.")

	flag.BoolVar(
		&generatorOptions.IncludePackageInTags, "include_package_in_tags", generatorOptions.IncludePackageInTags,
		"whether or not to include the proto package in the service name used in the operation tags.")

	flag.BoolVar(
		&generatorOptions.DisableServiceTags, "disable_service_tags", generatorOptions.DisableServiceTags,
		"if set, disables generation of service tags in operations. This is useful to avoid exposing backend gRPC service names.")

	flag.Var(
		&generatorOptions.SchemaNamingStrategy, "schema_naming_strategy",
		"controls the name of OpenAPI schemas. Use 'fqn' to use full name, 'simple' to use the shortest unique name"+
			" and 'simple+version' to include a version prefix when one is available (e.g., v1alpha1Message).")

	flag.BoolVar(
		&generatorOptions.UseGoTemplate, "use_go_templates", generatorOptions.UseGoTemplate,
		"if enabled, tags, titles, summaries, and links can use go templates. Refer to documentation for available values.")

	flag.Var(
		&generatorOptions.GoTemplateArgs, "go_template_args",
		"comma-separated assignment of Go template args. Example: a=b,c=d")

	flag.BoolVar(
		&generatorOptions.IgnoreComments, "ignore_comments", generatorOptions.IgnoreComments,
		"if set, proto documentation and comments get ignored completely.")

	flag.BoolVar(
		&generatorOptions.RemoveInternalComments, "remove_internal_comments", generatorOptions.RemoveInternalComments,
		"if set, any string wrapped in (-- and --) gets excluded.")

	flag.BoolVar(
		&generatorOptions.DisableDefaultErrors, "disable_default_errors", generatorOptions.DisableDefaultErrors,
		"if set, default error response does not get generated. Useful when custom error structure is used.")

	flag.BoolVar(
		&generatorOptions.DisableDefaultResponses, "disable_default_responses", generatorOptions.DisableDefaultResponses,
		"if set, default success response does not get generated. Useful when non 200 status codes are needed.")

	flag.BoolVar(
		&generatorOptions.UseEnumNumbers, "use_enum_numbers", generatorOptions.UseEnumNumbers,
		"if set, enums in the OpenAPI use the numerical values instead of string values.")

	flag.StringVar(
		&generatorOptions.GlobalOpenAPIConfigFile, "openapi_config", generatorOptions.GlobalOpenAPIConfigFile,
		"if set, this config file gets used as a top-level config for all proto files and services."+
			" One can use one config file for both gRPC and OpenAPI configs. By default if gateway config is specified,"+
			" unless '-' is used for openapi_config, the same gets used for OpenAPI configs as well.")

	flag.StringVar(
		&generatorOptions.OpenAPIConfigFilePattern, "openapi_config_pattern", generatorOptions.OpenAPIConfigFilePattern,
		"openapi file pattern (without the extension segment) that gets used to try and load an OpenAPI config file"+
			" for each proto file containing service definitions. yaml, yml and finally json file extensions will be tried.")

	flag.StringVar(
		&generatorOptions.OpenAPISeedFile, "openapi_seed_file", generatorOptions.OpenAPISeedFile,
		"if set, this OpenAPI file (yaml or json) gets used as a template and will get merged with the generated files."+
			" Useful to set values unavailable in the OpenAPI generation configs or to repeat document"+
			" values in all generated files.")

	flag.BoolVar(
		&generatorOptions.OmitEnumDefaultValue, "omit_enum_default_value", generatorOptions.OmitEnumDefaultValue,
		"if set, excludes the default value for all enums.")

	flag.Var(
		&generatorOptions.VisibilitySelectors, "visibility_selectors",
		"comma-separated list of included visibility labels. Example: INTERNAL,PARTNERS")

	flag.BoolVar(
		&generatorOptions.MergeWithOverwrite, "merge_with_overwrite", generatorOptions.MergeWithOverwrite,
		"when this option is enabled, arrays get overwritten instead of appended.")

	flag.BoolVar(
		&generatorOptions.OmitEmptyFiles, "omit_empty_files", generatorOptions.OmitEmptyFiles,
		"when enabled, OpenAPI documents that do not contain at least one generated schema/path get skipped.")

	flag.Var(
		&generatorOptions.FieldRequiredMode, "field_required_mode",
		"can be used to automatically mark fields as required. 'disabled' (default) does not automatically"+
			"mark any field, 'not_optional' marks any field that is not labled as optional as required and"+
			"'not_optional_scalar' is similar to the previous mode but only for scalar types (not message).")

	flag.Var(
		&generatorOptions.FieldNullableMode, "field_nullable_mode",
		"can be used to generate nullable OpenAPI fields using 'anyOf' or type array for scalar types."+
			" 'disabled' does not generate nullable fields at all,"+
			" 'optional' adds generates nullable fields when proto3 optional label is used"+
			" and 'not_required' adds nullable field when a field is not explicitly marked as required and can be null.")

	flag.BoolVar(
		&generatorOptions.LocalPackageMode, "local_package_mode", generatorOptions.LocalPackageMode,
		"if enabled, limits each config file (save for the global config) to the local proto package.")

	return &generatorOptions
}
