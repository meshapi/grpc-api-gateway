
# --8<-- [start:AdditionalEndpointBinding]
### AdditionalEndpointBinding

AdditionalEndpointBinding is an additional gRPC method - HTTP endpoint binding specification.

| <div style="width:120px">Field Name</div> | Type | Description |
| --- | --- | --- |
| `get` | string ([RoutePattern](#routepattern)) |  |
| `put` | string ([RoutePattern](#routepattern)) |  |
| `post` | string ([RoutePattern](#routepattern)) |  |
| `delete` | string ([RoutePattern](#routepattern)) |  |
| `patch` | string ([RoutePattern](#routepattern)) |  |
| `custom` | [CustomPattern](#custompattern)  | custom can be used for custom HTTP methods.<br><br>Not all HTTP methods are supported in OpenAPI specification, however and will not be included in the<br>generated OpenAPI document. |
| `body` | string  | body is a request message field selector that will be read via HTTP body.<br><br>'*' indicates that the entire request message gets decoded from the body.<br>An empty string indicates that no part of the request gets decoded from the body.<br><br>NOTE: Not all methods support HTTP body. |
| `response_body` | string  | response_body is a response message field selector that will be written to HTTP response.<br><br>'*' or an empty string indicates that the entire response message gets encoded. |
| `query_params` | [QueryParameterBinding](#queryparameterbinding)  | query_params are explicit query parameter bindings that can be used to rename<br>or ignore query parameters. |
| `disable_query_param_discovery` | bool  | disable_query_param_discovery can be used to avoid auto binding query parameters. |
| `stream` | [StreamConfig](#streamconfig)  | stream holds configurations for streaming methods. |
# --8<-- [end:AdditionalEndpointBinding]
# --8<-- [start:CustomPattern]
### CustomPattern

CustomPattern describes an HTTP pattern and custom method.

| <div style="width:120px">Field Name</div> | Type | Description |
| --- | --- | --- |
| `method` | string  | method is the custom HTTP method. |
| `path` | string  | path is the HTTP path pattern. |
# --8<-- [end:CustomPattern]
# --8<-- [start:EndpointBinding]
### EndpointBinding

EndpointBinding represents an HTTP endpoint(s) to gRPC method binding.

| <div style="width:120px">Field Name</div> | Type | Description |
| --- | --- | --- |
| `selector` | string  | selector is a dot-separated gRPC service method selector.<br><br>If the selector begins with `~.`, the current proto package will be added to the beginning<br>of the path. For instance: `~.MyService`. Since no proto package can be deduced in the global<br>config file, this alias cannot be used in the global config file.<br><br>If the selector does not begin with `~.`, it will be treated as a fully qualified method name (FQMN).<br><br>NOTE: In proto annotations, this field gets automatically assigned, thus it is only applicable in configuration files. |
| `get` | string ([RoutePattern](#routepattern)) | get defines route for a GET HTTP endpoint. |
| `put` | string ([RoutePattern](#routepattern)) | put defines route for a PUT HTTP endpoint. |
| `post` | string ([RoutePattern](#routepattern)) | post defines route for a POST HTTP endpoint. |
| `delete` | string ([RoutePattern](#routepattern)) | delete defines route for a DELETE HTTP endpoint. |
| `patch` | string ([RoutePattern](#routepattern)) | patch defines route for a PATCH HTTP endpoint. |
| `custom` | [CustomPattern](#custompattern)  | custom can be used for custom HTTP methods.<br><br>Not all HTTP methods are supported in OpenAPI specification and will not be included in the<br>generated OpenAPI document. |
| `body` | string  | body is a request message field selector that will be read via HTTP body.<br><br>`*` indicates that the entire request message gets decoded from the body.<br>An empty string (default value) indicates that no part of the request gets decoded from the body.<br><br>NOTE: Not all methods support HTTP body. |
| `response_body` | string  | response_body is a response message field selector that will be written to HTTP response.<br><br>`*` or an empty string indicates that the entire response message gets encoded. |
| `query_params` | [QueryParameterBinding](#queryparameterbinding)  | query_params are explicit query parameter bindings that can be used to rename<br>or ignore query parameters. |
| `additional_bindings` | [AdditionalEndpointBinding](#additionalendpointbinding)  | additional_bindings holds additional bindings for the same gRPC service method. |
| `disable_query_param_discovery` | bool  | disable_query_param_discovery can be used to avoid auto binding query parameters.<br><br>Default: `false` |
| `stream` | [StreamConfig](#streamconfig)  | stream holds configurations for streaming methods. |
# --8<-- [end:EndpointBinding]
# --8<-- [start:GatewaySpec]
### GatewaySpec

GatewaySpec holds gRPC gateway configurations.

| <div style="width:120px">Field Name</div> | Type | Description |
| --- | --- | --- |
| `endpoints` | [EndpointBinding](#endpointbinding)  | endpoints hold a series of endpoint binding specs. |
# --8<-- [end:GatewaySpec]
# --8<-- [start:QueryParameterBinding]
### QueryParameterBinding

QueryParameterBinding describes a query parameter to request message binding.

| <div style="width:120px">Field Name</div> | Type | Description |
| --- | --- | --- |
| `selector` | string  | selector is a dot-separated path to the request message's field. |
| `name` | string  | name is the name of the HTTP query parameter that will be used. |
| `ignore` | bool  | ignore avoids reading this query parameter altogether (default: false). |
# --8<-- [end:QueryParameterBinding]
# --8<-- [start:StreamConfig]
### StreamConfig

StreamConfig sets the behavior of the HTTP server for gRPC streaming methods.

| <div style="width:120px">Field Name</div> | Type | Description |
| --- | --- | --- |
| `disable_websockets` | bool  | disable_websockets indicates whether or not websockets are allowed for this method.<br>The client must still ask for a connection upgrade. |
| `disable_sse` | bool  | disable_sse indicates whether or not server-sent events are allowed.<br><br>see: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events<br><br>SSE is only used when Accept-Type from the request includes MIME type text/event-stream. |
| `disable_chunked_transfer` | bool  | disable_chunked indicates whether or not chunked transfer encoding is allowed.<br><br>NOTE: Chunked transfer encoding is disabled in HTTP/2 so this option will only be available if the request<br>is HTTP/1. |
# --8<-- [end:StreamConfig]
