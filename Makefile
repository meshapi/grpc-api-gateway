install:
	@go install ./codegen/cmd/protoc-gen-grpc-api-gateway/
	@go install ./codegen/cmd/protoc-gen-openapiv3/

docs-run:
	docker run --rm -it -p 8000:8000 -v ${PWD}:/docs squidfunk/mkdocs-material:9.1
