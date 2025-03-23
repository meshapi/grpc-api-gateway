# Intro

While gRPC API Gateway offers extensive customization and details,
this quick step-by-step guide aims to show you the fastest way to generate a reverse proxy for your gRPC service.

For a comprehensive reference of all features and configurations,
see [Reference](/grpc-api-gateway/reference/intro).

## Prerequisites

Before we start coding, ensure that you have installed the gRPC API Gateway protobuf plug-in.
Refer to the [Installation](/grpc-api-gateway/installation) section for detailed instructions.

gRPC API Gateway is a `protoc` plug-in, similar to the Go code generator. The tool relies on the generated code from both the [Go](https://protobuf.dev/reference/go/go-generated/) and [Go gRPC](https://grpc.io/docs/languages/go/quickstart/) plug-ins.

To install the Go and Go gRPC plug-ins, execute the following commands:

```sh
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
