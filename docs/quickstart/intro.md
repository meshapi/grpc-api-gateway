# Intro

There is a lot of customization and details to gRPC API Gateway, however this quick step by step guide aims to show
you the quickest way to generate a reverse proxy for your gRPC service.

See [Reference](/grpc-api-gateway/reference/intro) for a complete reference of all features and configurations.

## Prerequisites

Before we get to coding, ensure that you have installed the gRPC API Gateway protobuf plug-in.
Refer to [Installation](/grpc-api-gateway/installation) section for related instructions.

gRPC API Gateway is a `protoc` plug-in, similar to the Go code generator and the tool
relies on the generated code from both [Go](https://protobuf.dev/reference/go/go-generated/)
and [Go gRPC](https://grpc.io/docs/languages/go/quickstart/) plug-ins.

In order to install Go and Go gRPC plug-ins, execute the following commands:

```sh
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
