module github.com/meshapi/grpc-api-gateway/codegen

go 1.19.0

replace github.com/meshapi/grpc-api-gateway => ..

require (
	dario.cat/mergo v1.0.0
	golang.org/x/text v0.14.0
	golang.org/x/tools v0.18.0
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17
	google.golang.org/grpc v1.61.1
	google.golang.org/protobuf v1.32.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	golang.org/x/mod v0.15.0 // indirect
	google.golang.org/genproto v0.0.0-20231106174013-bbf56f31fb17 // indirect
)
