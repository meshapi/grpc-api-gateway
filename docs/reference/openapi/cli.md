# Plug-in Options

The _OpenAPI v3_ protoc plug-in offers a variety of configurable options. The table below provides a comprehensive list of these options along with their descriptions and default values.

Note that there are some overlaps between gRPC options and OpenAPI options. To ensure the OpenAPI plug-in generates an accurate OpenAPI document that matches your gRPC API Gateway, it is important to use the same options in both the gRPC API Gateway and OpenAPI v3 plug-ins.

!!! example
    If you use `allow_delete_body` in the gRPC API Gateway, ensure the same value is set for the OpenAPI v3 plug-in. This alignment guarantees that the generated OpenAPI document accurately reflects the gRPC API's behavior.

| Option      | Description                               | Default |
| --- | --- | --- |
| allow_delete_body | Allows HTTP DELETE methods to include a body if explicitly specified. | `false` |
| allow_patch_feature | Enables the use of the PATCH feature with update masks (`google.protobuf.FieldMask`). | `true` |
| disable_default_errors | When enabled, the default error response is not generated. This is useful if you are using a custom error structure. | `false` |
| disable_default_responses | When enabled, the default success response is not generated. This is particularly useful when you need to specify non-200 status codes. | `false` |
| disable_service_tags | When enabled, prevents the generation of service tags in operations. This helps to avoid exposing backend gRPC service names. | `false` |
| field_name_mode | Controls the naming convention of fields in the OpenAPI schemas. Allowed values are `proto` for using the original proto field names and `json` for using camelCase JSON names. | `json` |
| field_nullable_mode | Configures the generation of nullable OpenAPI fields for scalar types using `anyOf` or type array. Options include: `disabled` (no nullable fields generated), `optional` (nullable fields generated when the proto3 optional label is used), and `not_required` (nullable fields generated when a field is not explicitly marked as required and can be null). | `optional` |
| field_required_mode | Configures the automatic marking of fields as required. Options include: `disabled` (default) does not automatically mark any field as required, `non_optional` marks any field that is not labeled as optional as required, and `non_optional_scalar` marks only scalar types (not message types) that are not labeled as optional as required. | `disabled` |
| config_search_path | Specifies the directory (relative or absolute) from the current working directory that contains the gateway config files. See [Search Path](/grpc-api-gateway/reference/configuration/#search-path) for more information. | `.` |
| gateway_config | Path to the global gateway config file that is loaded first. This file can contain bindings for any service. | No default |
| gateway_config_pattern | Pattern (excluding the extension) used to load a gateway config file for each proto file. The extensions `.yaml`, `.yml`, and `.json` will be tried in that order. See [Filename Pattern](/grpc-api-gateway/reference/configuration/#filename-pattern) for more details. | `{{.Path}}_gateway` |
| openapi_config | If set, this configuration file is used as a top-level config for all proto files and services. You can use a single config file for both gRPC and OpenAPI configurations. By default, if a gateway config is specified, and unless '-' is used for openapi_config, the same config file will be used for OpenAPI configurations as well. | No default |
| openapi_config_pattern | Specifies the pattern (excluding the extension) used to load an OpenAPI config file for each proto file containing service definitions. The extensions `.yaml`, `.yml`, and `.json` will be tried in that order. | `{{ .Path }}_gateway` |
| openapi_seed_file | If set, this OpenAPI file (YAML or JSON) is used as a template and merged with the generated files. This is useful for setting values not available in the OpenAPI generation configs or for repeating document values across all generated files. | No default |
| generate_unbound_methods | Includes unannotated RPC methods in the proxy. Methods without explicit HTTP bindings will default to POST with the route pattern `/<grpc-service>/<method>`. | `false` |
| use_go_templates | When enabled, allows the use of Go templates for tags, titles, summaries, and links. Refer to the documentation for available template values. | `false` |
| go_template_args | Allows the assignment of Go template arguments in a comma-separated format. Example: `a=b,c=d`. | No default |
| ignore_comments | When enabled, all proto documentation and comments are completely ignored. | `false` |
| remove_internal_comments | When enabled, excludes any string wrapped in `(--` and `--)` from the generated output. | `false` |
| include_package_in_tags | Specifies whether to include the proto package name in the service name used in the operation tags. | `false` |
| operation_id_mode | Controls the format of operation IDs in the OpenAPI document. Options are: `service+method` for `<Service>_<Method>`, `method` for just the method name, and `fqn` for the fully qualified name. The default is `service+method`. | `service+method` |
| include_services_only | When enabled, generates only the bound service methods and the models required to support them. | `false` |
| local_package_mode | When enabled, restricts each configuration file (except the global config) to the local proto package. | `false` |
| log_file | If specified, the plug-in writes all logs to this file. | No default |
| log_level | Sets the log level. Available levels: `warning`, `info`, `trace`, and `silent`. | `warning` |
| merge_with_overwrite | When enabled, this option causes arrays to be overwritten rather than appended. | `true` |
| omit_empty_files | When enabled, skips the generation of OpenAPI documents that do not contain at least one schema or path. | `false` |
| omit_enum_default_value | When enabled, omits the default value for all enum fields in the generated OpenAPI document. | `false` |
| use_enum_numbers | When enabled, enums in the OpenAPI document will use their numerical values instead of string representations. | `false` |
| repeated_path_param_separator | Configures how repeated fields should be split. Allowed values are `csv`, `pipes`, `ssv`, and `tsv`. | `csv` |
| warn_on_unbound_methods | Emits a warning message if an RPC method has no mapping. | `false` |
| warn_on_broken_selectors | When enabled, reduces the severity of unrecognized selectors in configuration files to a warning level in the logs. | `false` |
| schema_naming_strategy | Controls the naming convention for OpenAPI schemas. Options include: `fqn` for using the fully qualified name, `simple` for using the shortest unique name, and `simple+version` for including a version prefix when available (e.g., `v1alpha1Message`). | `simple` |
| visibility_selectors | A comma-separated list of visibility labels to include. Example: `INTERNAL,PARTNERS`. When empty, all methods are selected. | No default |
| output_mode | Determines how the OpenAPI definitions are organized in the output. Options are: `merge` to combine all definitions into a single file, `proto` to generate one file per proto file, and `service` to create a separate document for each gRPC service. | `proto` |
| output_filename | Applicable only when using the `merge` output mode. This option sets the filename for the generated OpenAPI document. | `apidocs` |
| output_format | Specifies the format of the generated output. Allowed values are `yaml` and `json`. | `json` |
