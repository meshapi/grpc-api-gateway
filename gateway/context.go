package gateway

import "strings"

// malformedHTTPHeaders lists the headers that the gRPC server may reject outright as malformed.
// See https://github.com/grpc/grpc-go/pull/4803#issuecomment-986093310 for more context.
var malformedHTTPHeaders = map[string]struct{}{
	"connection": {},
}

// isMalformedHTTPHeader checks whether header belongs to the list of
// "malformed headers" and would be rejected by the gRPC server.
func isMalformedHTTPHeader(header string) bool {
	_, isMalformed := malformedHTTPHeaders[strings.ToLower(header)]
	return isMalformed
}
