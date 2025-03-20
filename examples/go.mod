module github.com/meshapi/grpc-api-gateway/examples

go 1.22.0
toolchain go1.24.1

replace (
	github.com/meshapi/grpc-api-gateway => ..
	github.com/meshapi/grpc-api-gateway/websocket/wrapper/gorillawrapper => ../websocket/wrapper/gorillawrapper
)

require (
	github.com/google/go-cmp v0.6.0
	github.com/gorilla/websocket v1.5.1
	github.com/meshapi/grpc-api-gateway v0.0.0-00010101000000-000000000000
	github.com/meshapi/grpc-api-gateway/websocket/wrapper/gorillawrapper v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.61.1
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	golang.org/x/net v0.36.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
)
