# Configuration Reference

To define and bind HTTP endpoints to gRPC methods, you can use either
configuration files or proto annotations directly within the proto files.
For more details, refer to the [Configuration](/grpc-api-gateway/reference/configuration) documentation.

Gateway configuration files should include the following object
(`GatewayConfig`) under the `gateway` key:

| <div style="width:120px">Field Name</div> | Type                                  | Description                     |
| ----------------------------------------- | ------------------------------------- | ------------------------------- |
| `endpoints`                               | [[EndpointBinding](#endpointbinding)] | List of all gRPC-HTTP bindings. |

!!! example

    ```yaml
    gateway:
      endpoints:
        - selector: "~.MyService.MyMethod"
          get: "/route"
    ```

--8<-- "templates/gateway.md:EndpointBinding"

#### HTTP Method & Route

Each [EndpointBinding](#endpointbinding) object can define multiple HTTP bindings to a single gRPC method using
the `additional_bindings` property. You must specify exactly one of the following fields to define the HTTP method:
`get`, `post`, `put`, `patch`, `delete`, or `custom`.

To use an HTTP method not listed, utilize the `custom` property.

!!! warning
    Any HTTP method can be used for the gRPC gateway. However, OpenAPI specification supports only
    a limited set of HTTP methods. Unsupported methods will not appear in the generated OpenAPI documents.

When using the `custom` field, the value must be a [CustomPattern](#custompattern).
For other fields, the value should be a [RoutePattern](#routepattern).

!!! example
    === "Configuration"
        ```yaml title="sound_gateway.yaml" linenums="1" hl_lines="3"
        gateway:
          endpoints:
            - post: "/echo"
              selector: "~.SoundService.Echo"
        ```

    === "Proto Annotations"
        ```proto title="sound.proto" linenums="1" hl_lines="3-5"
        service SoundService {
            rpc Echo(EchoRequest) returns (EchoResponse) {
                option (meshapi.gateway.http) = {
                    post: "/echo"
                };
            }
        }
        ```

#### RoutePattern

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

    Nested fields are supported so `/path/{name}/{nested.field}` is valid.

#### Wildcard

If you want a field to contain all segments, including slashes, you can use the `{<selector>=*}` pattern.

!!! example

    ```proto linenums="1"
    message Request {
        string file_path = 1;
    }
    ```

    Route pattern `/path/{file_path=*}` would match `/path/a/b/c/d/e.jpg` and capture `file_path` as `/a/b/c/d/e.jpg`.

### AdditionalEndpointBinding

This object is similar to [EndpointBinding](#endpointbinding) excluding the `additional_bindings` key.

!!! example
    === "Configuration"
        ```yaml title="sound_gateway.yaml" linenums="1" hl_lines="5-11"
        gateway:
          endpoints:
            - get: "/echo"
              selector: "~.SoundService.Echo"
              additional_endpoints:
                - get: "/another-route"
                - post: "/echo-with-post"
                  body: "*"
                - custom:
                    path: "/echo"
                    method: "LOG"
        ```

    === "Proto Annotations"
        ```proto title="sound.proto" linenums="1" hl_lines="5-11"
        service SoundService {
            rpc Echo(EchoRequest) returns (EchoResponse) {
                option (meshapi.gateway.http) = {
                    get: "/echo",
                    additional_endpoints: [
                      {get: "/another-route"},
                      {post: "/echo-with-post", body: "*"},
                      {
                        custom: {method: "LOG", path: "/echo"}
                      }
                    ]
                };
            }
        }
        ```

--8<-- "templates/gateway.md:CustomPattern"

!!! example
    === "Configuration"
        ```yaml title="sound_gateway.yaml" linenums="1" hl_lines="3-5"
        gateway:
          endpoints:
            - custom:
                method: "TRACE"
                path: "/echo"
              selector: "~.SoundService.Echo"
        ```

    === "Proto Annotations"
        ```proto title="sound.proto" linenums="1" hl_lines="4-7"
        service SoundService {
            rpc Echo(EchoRequest) returns (EchoResponse) {
                option (meshapi.gateway.http) = {
                    custom: {
                        method: "TRACE",
                        path: "/echo"
                    }
                };
            }
        }
        ```

!!! warning
    Any HTTP method can be used and will function correctly in the gateway.
    However, methods not supported by OpenAPI will be excluded from the generated documentation.

--8<-- "templates/gateway.md:QueryParameterBinding"

By default, any field in the request proto message that is not bound to the HTTP body or path parameters will be automatically bound to query parameters.

You can explicitly bind one or more fields to query parameters by specifying the proto message selector and the desired query parameter name. Alternatively, you can use `ignore` to exclude specific fields from being bound to query parameters.

!!! example
    Consider the following request message:
    ```proto linenums="1"
    message EchoOptions {
        int32 delay = 1;
        bool lower_case = 2;
    }

    message EchoRequest {
        EchoOptions options = 1;
        string message = 2;
    }
    ```

    === "Configuration"
        ```yaml title="sound_gateway.yaml" linenums="1" hl_lines="5-11"
        gateway:
          endpoints:
            - post: "/echo"
              selector: "~.SoundService.Echo"
              query_parameters:
                - selector: "message"
                  name: "msg"
                - selector: "options.lower_case" # (1)!
                  name: "lower"
                - selector: "options.delay"
                  ignore: true # (2)!
        ```

        1. Defining aliases for long or nested fields can simplify query parameters. For example, using `lower` instead of `options.lower_case`.
        2. Setting `ignore` to true prevents the `options.delay` proto field from being bound to any query parameter.

    === "Proto Annotations"
        ```proto title="sound.proto" linenums="1" hl_lines="5-9"
        service SoundService {
            rpc Echo(EchoRequest) returns (EchoResponse) {
                option (meshapi.gateway.http) = {
                    post: "/echo",
                    query_parameters: [
                        {selector: "message", name: "msg"},
                        {selector: "options.lower_case", name: "lower"}, // (1)!
                        {selector: "options.delay", ignore: true} // (2)!
                    ]
                };
            }
        }
        ```

        1. Defining aliases for long or nested fields can simplify query parameters. For example, using `lower` instead of `options.lower_case`.
        2. Setting `ignore` to true prevents the `options.delay` proto field from being bound to any query parameter.

    In this example, in the HTTP request, you can use `msg` and `lower` query parameters directly:

    `/echo?msg=something&lower=true`

!!! info
    Defining aliases replaces the default auto-binding names. In the example above, using `msg` as an alias for the proto field `message` means only the query parameter `msg` will be bound to the proto field `message`. If you want to keep the original name as well, you can define multiple aliases for the same selector.

--8<-- "templates/gateway.md:StreamConfig"

!!! example
    Imagine an event streaming endpoint that continuously sends events to the client. Using chunked transfer for this is not ideal due to timeout constraints. However, using *SSE* (Server-Sent Events) or *WebSockets* is perfectly valid and recommended.
    === "Configuration"
        ```yaml title="notification_gateway.yaml" linenums="1" hl_lines="5-6"
        gateway:
          endpoints:
            - post: "/notify"
              selector: "~.NotificationService.Notify"
              stream:
                disable_chunked_transfer: true
        ```

    === "Proto Annotations"
        ```proto title="notification.proto" linenums="1" hl_lines="5-7"
        service NotificationService {
            rpc Notify(NotifyRequest) returns (stream NotifyResponse) {
                option (meshapi.gateway.http) = {
                    get: "/events",
                    stream: {
                        disable_chunked_transfer: true
                    }
                };
            }
        }
        ```
