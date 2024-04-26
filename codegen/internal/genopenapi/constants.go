package genopenapi

const (
	extYAML = "yaml"
	extJSON = "json"
)

const (
	fqmnAny      = ".google.protobuf.Any"
	fqmnHTTPBody = ".google.api.HttpBody"
)

const openAPIOutputSuffix = ".openapi"

const (
	commentInternalOpen  = "(--"
	commentInternalClose = "--)"
)

const refPrefix = "#/components/schemas/"

const defaultSuccessfulResponse = "a successful response."
const (
	mimeTypeJSON = "application/json"
	mimeTypeSSE  = "text/event-stream"
)
const httpStatusOK = "200"
const httpStatusDefault = "default"
const rpcStatusProto = ".google.rpc.Status"
const streamingInputDescription = " (streaming inputs)"
const streamingResponsesDescription = " (streaming responses)"
const headerTransferEncoding = "Transfer-Encoding"
const encodingChunked = "chunked"
