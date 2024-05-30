# OpenAPI Documentation

OpenAPI documentation is already generated for our service.
However, we might like to change a few things in this document.

In this page we will customize our OpenAPI document for this `Echo` service.

Refer to [Reference](/grpc-api-gateway/reference/intro) to learn more about all of the customizations.

Similar to defining HTTP bindings, you can opt to use the annotations and directly define options in the proto
files or use configuration files.


=== "Using configurations"

    ```yaml title="echo_service_gateway.yaml" linenums="1"
    # ... omitted gateway spec for brevity.

    openapi:
      document: #(1)!
        info:
          title: 'Echo Service'
          version: 'v0.0.1-alpha1'

      services:
        - selector: '~.EchoService' #(2)!
          methods:
            Echo: #(3)!
              external_docs:
                url: 'http://meshapi.github.com/grpc-api-gateway'
                description: 'Even more documentation!'

      messages:
        - selector: 'EchoRequest' #(4)!
          fields:
            'text':
              description: "Text is the input text"
              max_length: "24"
    ```

    1. `document` allows you to customize the resulting OpenAPI document for this proto file. We will use it to set
       `title` and `version` here.
    2. Similar to the _selector_ we defined in the HTTP bindings, this is a dotted path to the service and `~` resolves
       to the current proto package.
    3. Defining `external_docs` in the OpenAPI document only for the HTTP endpoints that are bound to the `Echo` gRPC
       method.
    4. Similar to the _selector_ we had to define for the service, however here we point to a message to customize
       `schema` for this proto message. In this case, to set custom description and include additional validation
       details.

=== "Using proto extensions"

    ```proto title="echo_service.proto" linenums="1" hl_lines="8-13 16-21 42-47"
    syntax = "proto3";

    package echo;

    import "meshapi/gateway/annotations.proto";

    option go_package = "demo/echo";
    option (meshapi.gateway.openapi_doc) = { //(1)!
        info: {
            title: 'Echo Service',
            version: 'v0.0.1-alpha1'
        }
    };

    message EchoRequest {
        string text = 1 [
            (meshapi.gateway.openapi_field) = { //(2)!
                description: 'Text is the input text',
                max_length: '24'
            }
        ];
        bool capitalize = 2;
    }

    message EchoResponse {
        string text = 1;
    }

    service EchoService {
        // Echo returns the received text and make it louder too!
        rpc Echo(EchoRequest) returns (EchoResponse) {
            option (meshapi.gateway.http) = {
                get: '/echo/{text}'
                additional_bindings: [
                  {
                    post: '/echo',
                    body: '*'
                  }
                ]
            };

            option (meshapi.gateway.openapi_operation) = { //(3)!
                external_docs: {
                    url: 'http://meshapi.github.com/grpc-api-gateway',
                    description: 'Even more documentation!'
                }
            };
        };
    }
    ```

    1. This option allows you to customize the resulting OpenAPI document for this proto file. We will use it to set
       `title` and `version` here.
    2. This option allows you to customize the _schema_ generated for this field in the resulting OpenAPI document.
       Here we add custom description and extra validation details.
    3. This option can be used to customize _operation_ for HTTP endpoints bound to this gRPC method.
        Here, defining `external_docs` in the OpenAPI document only for the HTTP endpoints that are
        bound to the `Echo` gRPC method.


Now you can re-generate using either `Buf` or `protoc` directly and notice the changes in the generated OpenAPI file.

That's it for our quick guide!
