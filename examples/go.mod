module github.com/meshapi/grpc-rest-gateway/examples

go 1.19.0

replace (
	github.com/meshapi/grpc-rest-gateway => ..
	github.com/meshapi/grpc-rest-gateway/websocket/wrapper/gorillawrapper => ../websocket/wrapper/gorillawrapper
)

require (
	github.com/google/go-cmp v0.6.0
	github.com/gorilla/websocket v1.5.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/meshapi/grpc-rest-gateway v0.0.0-00010101000000-000000000000
	github.com/meshapi/grpc-rest-gateway/websocket/wrapper/gorillawrapper v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.61.1
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
)
