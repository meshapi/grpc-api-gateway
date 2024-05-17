# Add HTTP Bindings

## Define HTTP Bindings

Now we need to define the HTTP bindings for our gRPC service.
To do this, we need to define gRPC method to HTTP endpoint bindings.

Generally this can be either via gRPC API Gateway proto extensions or via a configuration file.

The way configurations get loaded can be heavily customized. Refer to [Reference](/reference) to learn more.
The default behavior is that for any proto file `file.proto`, files `file_gateway.yaml`, `file_gateway.yml` and `file_gateway.json` will be tried in order. If any file is available, it will be consumed.

It is a matter of personal preference to choose the method that works best for you and your project.
In this document, guidelines for both methods is provided.

!!! info
    You can utilize the `generate_unbound_methods` to automatically define HTTP bindings for the gRPC methods 
    that do not already have a defined binding. The path will be in form of `/proto.package.Service/Method`
    and the HTTP method will be `POST`.


Let's create two HTTP endpoints for the `Echo` method, one using `GET` and one using `POST` with your method of preference, either configuration files or adding proto annotations.

### 1. Using Configuration Files

We will create a new file with the pattern `<proto-filename>_gateway.yaml`. Since we used `echo_service.proto` in this demo, we will create `echo_service_gateway.yaml` file with the content below:

```yaml title="echo_service_gateway.yaml" linenums="1"
gateway:
  endpoints:
    - selector: '~.EchoService.Echo' # (1)!
      get: '/echo/{text}' # (2)!
      additional_bindings: # (3)!
        - post: '/echo'
          body: '*' # (4)!
```

1. `selector` is dot-separated path to the service method. `echo.EchoService.Echo` is the full path. `~` gets substituted with the proto pacakge from the related proto file and can be used to shorten the selector.
2. This line states that HTTP method is `GET` and the route is `/echo/{text}` where `text` is a path to a field in the request proto message.
3. Can be used to specify additional HTTP endpoint bindings.
4. `body` (default: `null`) can be used to specify which fields in the proto request message should be read from the HTTP body. In this case, `*` indicates that all fields in the proto message should be read from the HTTP body.

!!! tip
    If you've chosen to work with configuration files, consider installing a YAML or JSON extension for your editor.
    Files named according to the pattern `*_gateway.[yml|yaml|json]` utilize the API Gateway's configuration schema,
    providing auto-completion and in-editor documentation features.

### 2. Using proto extensions

In order to use the proto extensions, we need to first download and import the gRPC API Gateway annotations.

=== "Using Buf"

    Create another file named `buf.yaml` with the following content.

    ```yaml title="buf.yaml" linenums="1"
    version: v1
    deps:
      - "buf.build/meshapi/grpc-api-gateway"
    ```

    Now download the dependencies using:

    ```sh
    $ buf mod update
    ```

=== "Using protoc"

    Download and proto files from the `api/meshapi` directory in the
    [gRPC API Gateway git repository](https://github.com/meshapi/grpc-api-gateway/tree/main/api/meshapi/gateway)
    to a local directory named `meshapi`.

Let's modify our existing proto file with the following additions:

```proto title="echo_service.proto" hl_lines="5 21-29" linenums="1"
syntax = "proto3";

package echo;

import "meshapi/gateway/annotations.proto"; //(1)!

option go_package = "demo/echo";

message EchoRequest {
    string text = 1;
    bool capitalize = 2;
}

message EchoResponse {
    string text = 1;
}

service EchoService {
    // Echo returns the received text and make it louder too!
    rpc Echo(EchoRequest) returns (EchoResponse) {
        option (meshapi.gateway.http) = {
            get: '/echo/{text}' //(2)!
            additional_bindings: [ //(3)!
              {
                post: '/echo',
                body: '*' //(4)!
              }
            ]
        };
    };
}
```

1. This line imports the gRPC API Gateway proto annotations.
2. This line states that HTTP method is `GET` and the route is `/echo/{text}` where `text` is a path to a field in the request proto message.
3. Can be used to specify additional HTTP endpoint bindings.
4. `body` (default: `null`) can be used to specify which fields in the proto request message should be read from the HTTP body. In this case, `*` indicates that all fields in the proto message should be read from the HTTP body.

## Add HTTP Server

Now that we have defined HTTP bindings, we will need to regenerate the gateway code.

=== "Using Buf"

    ```sh
    $ buf generate
    ```

=== "Using protoc"
    ```sh
    $ protoc \
        --go_out=gen \
        --go-grpc_out=gen \
        --grpc-api-gateway_out=gen \
        --openapiv3_out=gen \
        echo_service.proto 
    ```

Using either method, we should now see a new file named `echo_service.pb.rgw.go`.

Now we need to get the `meshapi/grpc-api-gateway` module since it contains necessary types for our HTTP server:

```sh
$ go get github.com/meshapi/grpc-api-gateway
```

And finally, we go back to our `main.go` to add the HTTP server:

```go title="main.go" linenums="1" hl_lines="10 12 23-30 34 36-38"
package main

import (
	"context"
	"demo/gen/demo/echo"
	"log"
	"net"
	"strings"

	"github.com/meshapi/grpc-api-gateway/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ... removed service implementation for brevity.

func main() {
	listener, err := net.Listen("tcp", ":40000")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	gateway := gateway.NewServeMux()

	connection, err := grpc.NewClient( //(1)!
        ":40000",
        grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not dial gRPC server: %v", err)
	}

	server := grpc.NewServer()
	echo.RegisterEchoServiceServer(server, Service{})
	echo.RegisterEchoServiceHandler(context.Background(), gateway, connection) //(2)!

	go func() {
		log.Fatalln(http.ListenAndServe("0.0.0.0:4000", gateway))
	}()

	if err := server.Serve(listener); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
```

1. We create a gRPC connection to our own gRPC server. gRPC API Gateway uses this connection to communicate with the gRPC our services.
2. `RegisterEchoServiceHandler` is a generated function and will register the `EchoService` to the gateway mux.


## See it in action

Time to run the code and see it work:

```sh
$ go run .
```

and you should be able to send an HTTP request and get a response back:

```sh
$ curl http://localhost:4000/echo/greetings
```

You should get the following response back:

```json
{"text":"greetings"}
```
