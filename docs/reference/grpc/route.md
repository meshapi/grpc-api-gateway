# EndpointSpec

To define and bind HTTP endpoints to gRPC methods, you can use
either configuration files or proto annotations directly in the proto files.
See [Configuration](/grpc-api-gateway/reference/configuration) to learn more.

### EndpointSpec

Represents an HTTP endpoint(s) to gRPC method binding.

| <div style="width:110px">Field Name</div> | Type | Description |
| --- | --- | --- |
| [`<http-method>*`](#http-method-route) | [RoutePattern](#routepattern) | defines route for an HTTP method, method name can be `get`, `post`, `put`, `delete` or `patch`. |
| custom | [CustomPattern](#routepattern) | custom can be used for custom HTTP methods.<br>Not all HTTP methods are supported in OpenAPI specification, however and will not be included in the generated OpenAPI document.</br> |
| body | string | (Default: `''`)<br>request message field selector that will be read via HTTP body.</br>- `'*'` indicates that the entire request message gets decoded from the body.<br>- An empty string indicates that no part of the request gets decoded from the body.</br> |
| response_body | string | response message field selector that will be written to HTTP response.<br>`'*'` or an empty string indicates that the entire response message gets encoded.</br> |
| query_params | [QueryParameterBinding](#routepattern) | explicit query parameter bindings that can be used to rename or ignore query parameters. |
| additional_bindings | \[[AdditionalEndpointBinding](#routepattern)\] | additional HTTP bindings for the same gRPC method. |
| disable_query_param_discovery | boolean | disable_query_param_discovery can be used to avoid auto binding query parameters. |
| stream | [StreamConfig](#routepattern) | stream holds configurations for streaming methods. |

#### HTTP Method & Route

Each [EndpointSpec](#endpointspec) object can define multiple HTTP bindings to
one gRPC method via `additional_bindings` property.
HTTP method needs to be defined by specifying precisely one and only one of the
`get`, `post`, `put`, `patch`, `delete` or `custom` fields.

If `custom` field is used, the value must be a [CustomPattern](#custompattern).

For other fields, the value is a [RoutePattern](#routepattern).

!!! example
    === "Configuration"
        ```yaml title="echo_gateway.yaml" linenums="1" hl_lines="3"
        gateway:
          endpoints:
            - post: "/echo"
              selector: "~.SoundService.Echo"
        ```

    === "Proto Annotations"
        ```proto title="echo.proto" linenums="1" hl_lines="3-5"
        service SoundService {
            rpc Echo(EchoRequest) returns (EchoResponse) {
                option (meshapi.gateway.http) = {
                    post = "/echo"
                }
            }
        }
        ```

### AdditionalEndpointBinding

This object is nearly the same object as [EndpointBinding](#endpointspec_1) excluding the `additional_binding` key.

### RoutePattern

This is a string pattern that can contain parameters bound to the proto request fields. Message field selectors enclosed in curly braces get bound to the request message payload.

For instance:
`/path/{path.to.field}`

!!! example
    Consider the following request message:
    ```proto linenums="1"
    message NestedMessage {
        string field = 1;
    }

    message Request {
        NestedMessage nested = 1;
        string name = 2;
    }
    ```

    `/path/{name}` would bind to field `name` of the `Request` message.

    Nested fields are supported so `/path/{name}/{nested.field}` is perfectly valid.


Additionally, if you wanted a field to contain all segments including the slashes, you can use `{<selector>=*}` pattern.

!!! example

    ```proto linenums="1"
    message Request {
        string file_path = 1;
    }
    ```

    Route pattern `/path/{file_path=*}` would match `/path/a/b/c/d/e.jpg` and capture `file_path` as `/a/b/c/d/e.jpg`.
