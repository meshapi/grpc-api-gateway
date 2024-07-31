# Configuration Reference

To define and bind HTTP endpoints to gRPC methods, you can use
either configuration files or proto annotations directly in the proto files.
See [Configuration](/grpc-api-gateway/reference/configuration) to learn more.

Gateway configuration files accept the following object (`GatewayConfig`) under `gateway` key:

| <div style="width:120px">Field Name</div> | Type | Description |
| --- | --- | --- |
| `endpoints` | [[EndpointBinding](#endpointbinding)] | List of all gRPC-HTTP bindings. |


!!! example

    ```yaml
    gateway:
        endpoints:
            - selector: "~.MyService.MyMethod"
              get: "/route"
    ```

--8<-- "templates/gateway.md:EndpointBinding"

#### HTTP Method & Route

Each [EndpointBinding](#endpointbinding) object can define multiple HTTP bindings to
one gRPC method via `additional_bindings` property.
HTTP method needs to be defined by specifying precisely one and only one of the
`get`, `post`, `put`, `patch`, `delete` or `custom` fields.

If you would like to use an HTTP method that is not listed, you can use the `custom` property to use any HTTP method.

!!! warning
    You can use any HTTP method for the gRPC gateway. However, since OpenAPI specification only supports a limited set of HTTP methods, the unsupported methods do NOT get listed in the generated OpenAPI documents.

If the `custom` field is used, the value must be a [CustomPattern](#custompattern).
For other fields, the value is a [RoutePattern](#routepattern).

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
    Any method can be used and will work in the gateway.
    However, methods not recognized by OpenAPI will be skipped when generating the documentation.

--8<-- "templates/gateway.md:QueryParameterBinding"

By default, any fields in the request proto message not bound to the HTTP body
or path parameters are bound to query parameters.

You can bind one or more fields to query parameters by specifying
the proto message selector and the query parameter name or use `ignore` to avoid binding them at all.

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

        1. It can be helpful to define aliases for long or nested fields, in this case `lower` instead of `options.lower_case`.
        2. Setting `ignore` to true ignores binding `options.delay` proto field to **any** query parameter.

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

        1. It can be helpful to define aliases for long or nested fields, in this case `lower` instead of `options.lower_case`.
        2. Setting `ignore` to true ignores binding `options.delay` proto field to **any** query parameter.

    In this example, in the HTTP request, you can use `msg` and `lower` query parameters directly:

    `/echo?msg=something&lower=true`

!!! info
    Defining aliases overrides the default auto-binding names. In the example above, when using `msg` as an alias for
    the proto field `message`, only the query parameter `msg` will be bound to the proto field `message`. To retain the
    original name, you can define additional aliases for the same selector.

--8<-- "templates/gateway.md:StreamConfig"

!!! example
    Imagine there is an events streaming endpoint that keeps
    sending events to the client. For this, using chunked transfer is not ideal
    because of the time out constraints. Using *SSE* or *WebSockets* however is perfectly valid.
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
