# Plug-in Options

The _gRPC API Gateway_ protoc plug-in offers a variety of configurable options. The table below provides a comprehensive list of these options along with their descriptions and default values.

| Option      | Description                               | Default |
| --- | --- | --- |
| allow_delete_body | Allows HTTP DELETE methods to include a body if explicitly specified. | `false` |
| allow_patch_feature | Enables the use of the PATCH feature with update masks (`google.protobuf.FieldMask`). | `true` |
| config_search_path | Specifies the directory (relative or absolute) from the current working directory that contains the gateway config files. See [Search Path](/grpc-api-gateway/reference/configuration/#search-path) for more information. | `.` |
| gateway_config | Path to the global gateway config file that is loaded first. This file can contain bindings for any service. | No default |
| gateway_config_pattern | Pattern (excluding the extension) used to load a gateway config file for each proto file. The extensions `.yaml`, `.yml`, and `.json` will be tried in that order. See [Filename Pattern](/grpc-api-gateway/reference/configuration/#filename-pattern) for more details. | `{{.Path}}_gateway` |
| generate_local | __`Experimental`__ Generates code to directly use the server implementation instead of using gRPC clients. | `false` |
| generate_unbound_methods | Includes unannotated RPC methods in the proxy. Methods without explicit HTTP bindings will default to POST with the route pattern `/<grpc-service>/<method>`. | `false` |
| log_file | If specified, the plug-in writes all logs to this file. | No default |
| log_level | Sets the log level. Available levels: `warning`, `info`, `trace`, and `silent`. | `warning` |
| omit_package_doc | If true, no package comment will be included in the generated code. | `false` |
| register_func_suffix | Suffix used to construct names of generated `Register*<Suffix>` methods. | `Handler` |
| repeated_path_param_separator | Configures how repeated fields should be split. Allowed values are `csv`, `pipes`, `ssv`, and `tsv`. | `csv` |
| request_context | Determines whether to use the HTTP request's context. | `true` |
| standalone | Generates a standalone gateway package that imports the target service package. | `false` |
| warn_on_unbound_methods | Emits a warning message if an RPC method has no mapping. | `false` |
