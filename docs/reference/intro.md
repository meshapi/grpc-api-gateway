# Introduction

This documentation provides comprehensive guidelines and code examples
for utilizing OpenAPI v3.1 and gRPC API Gateway plug-ins.

The project includes two essential `protoc` plug-ins:

1. **`protoc-gen-grpc-api-gateway`**: This plug-in generates a reverse proxy in Go,
   leveraging the code produced by the [Go](https://protobuf.dev/reference/go/go-generated/)
   and [Go gRPC](https://grpc.io/docs/languages/go/quickstart/) plug-ins.
   The reverse proxy functions as an HTTP handler that converts HTTP requests into gRPC.

2. **`protoc-gen-openapiv3`**: This plug-in creates an OpenAPI v3.1 document for the reverse proxy HTTP server.
   As OpenAPI v3.1 is tailored for RESTful HTTP APIs, this plug-in currently does not support WebSockets,
   which may result in the exclusion of some streaming endpoints from the generated document.
