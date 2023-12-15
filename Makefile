grpc-rest-gateway:
	@go build -o out/ ./cmd/protoc-gen-grpc-rest-gateway/

install:
	@go install ./cmd/protoc-gen-grpc-rest-gateway/
