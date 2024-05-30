# Plug-in Options

_gRPC API Gateway_ protoc plug-in has a number of options. Table below lists them all.

| <div style="width:175px">Option</div>      | Description                               | <div style="width:130px">Default</div> |
| --- | --- | --- |
| `allow_delete_body` | By default, HTTP DELETE methods may not include a body unless explicitly specified. | `false` |
| `allow_patch_feature` | Determines whether to use PATCH feature involving update masks (using google.protobuf.FieldMask). | `true` |
| `config_search_path` | The gateway config search path is the directory (relative or absolute) from the current working directory that contains the gateway config files. See [Search Path](/grpc-api-gateway/reference/configuration/#search-path) for more information. | `.` |
| `gateway_config` | Specifies the path to the global gateway config file that is loaded first. This file can contain bindings for any service. | no default |
| `gateway_config_pattern` | The gateway file pattern (excluding the extension) used to load a gateway config file for each proto file. The extensions `.yaml`, `.yml`, and `.json` will be tried in that order. See [Filename Pattern](/grpc-api-gateway/reference/configuration/#filename-pattern) to learn more. | `{{.Path}}_gateway` |
| `generate_local` | __`Experimental`__ Generates code to directly use the server implementation instead of using gRPC clients | `false` |
| `generate_unbound_methods` | Determines whether unannotated RPC methods should be included in the proxy. Methods without explicit HTTP bindings will default to POST and will have the route pattern `/<grpc-service>/<method>`. | `false` |
| `log_file` | If specified, plug-in writes all logs to this file instead. | no default |
| `log_level` | Sets the log level, levels: `warning`, `info`, `trace` and `silent` | `warning` |
| `omit_package_doc` | If true, no package comment will be included in the generated code | `false` |
| `register_func_suffix` | Used to construct names of generated `Register*<Suffix>` methods. | `Handler` |
| `repeated_path_param_separator` | Configures how repeated fields should be split. Allowed values are `csv`, `pipes`, `ssv` and `tsv`. | `csv` |
| `request_context` | Determine whether to use HTTP request's context or not. | `true` |
| `standalone` | Generates a standalone gateway package, which imports the target service package | `false` |
| `warn_on_unbound_methods` | Emits a warning message if an RPC method has no mapping. | `false` |
