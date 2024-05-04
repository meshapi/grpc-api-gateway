package genopenapi

const (
	extYAML = "yaml"
	extJSON = "json"
)

const (
	fqmnAny       = ".google.protobuf.Any"
	fqmnHTTPBody  = ".google.api.HttpBody"
	fqmnFieldMask = ".google.protobuf.FieldMask"
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
const (
	httpStatusOK      = "200"
	httpStatusDefault = "default"

	rpcStatusProto                = ".google.rpc.Status"
	streamingInputDescription     = " (streaming inputs)"
	streamingResponsesDescription = " (streaming responses)"
	headerTransferEncoding        = "Transfer-Encoding"
	encodingChunked               = "chunked"

	fieldNameUpdateMask = "update_mask"
)
