all: gateway

# Generates gateway.md file.
gateway.md: template.md.tpl
	@protoc \
		--doc_out=. \
		--doc_opt=template.md.tpl,gateway.md \
		-I ../../api \
		meshapi/gateway/gateway.proto
