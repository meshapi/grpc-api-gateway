# Endpoint Spec

To define and bind HTTP endpoints to gRPC methods, you can use
either configuration files or proto annotations directly in the proto files.
See [Configuration](/grpc-api-gateway/reference/configuration) to learn more.

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


Additionally, if you want a field to contain all segments including slashes, you can use the `{<selector>=*}` pattern.

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

--8<-- "templates/gateway.md:QueryParameterBinding"

--8<-- "templates/gateway.md:StreamConfig"
