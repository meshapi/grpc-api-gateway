module github.com/meshapi/grpc-rest-gateway/codegen

go 1.19.0

replace github.com/meshapi/grpc-rest-gateway => ..

require (
	github.com/meshapi/grpc-rest-gateway v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.61.1
	google.golang.org/protobuf v1.32.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
)
