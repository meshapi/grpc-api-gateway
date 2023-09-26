package gateway

import (
	"context"
	"mime"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/meshapi/grpc-rest-gateway/gateway/internal/marshal"
	"go.starlark.net/lib/proto"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type DoubleArrayNotSortedOutYet int8

// ErrorHandlerFunc is the signature used to configure error handling.
type ErrorHandlerFunc func(context.Context, *ServeMux, Marshaler, http.ResponseWriter, *http.Request, error)

// StreamErrorHandlerFunc is the signature used to configure stream error handling.
type StreamErrorHandlerFunc func(context.Context, error) *status.Status

// RoutingErrorHandlerFunc is the signature used to configure error handling for routing errors.
type RoutingErrorHandlerFunc func(context.Context, *ServeMux, Marshaler, http.ResponseWriter, *http.Request, int)

// HeaderMatcherFunc checks whether a header key should be forwarded to/from gRPC context.
type HeaderMatcherFunc func(string) (string, bool)

// ForwardResponseFunc updates the outgoing gRPC request and the HTTP response.
type ForwardResponseFunc func(context.Context, http.ResponseWriter, proto.Message) error

// MetadataAnnotatorFunc updates the outgoing gRPC request context based on the incoming HTTP request.
type MetadataAnnotatorFunc func(context.Context, *http.Request) metadata.MD

// QueryParameterParser defines interface for all query parameter parsers
type QueryParameterParser interface {
	Parse(msg proto.Message, values url.Values, filter DoubleArrayNotSortedOutYet) error
}

type ServeMux struct {
	router *httprouter.Router

	// handlers maps HTTP method to a list of handlers.
	queryParamParser          QueryParameterParser
	forwardResponseOptions    []ForwardResponseFunc
	marshalers                marshal.Registry
	incomingHeaderMatcher     HeaderMatcherFunc
	outgoingHeaderMatcher     HeaderMatcherFunc
	metadataAnnotators        []MetadataAnnotatorFunc
	errorHandler              ErrorHandlerFunc
	streamErrorHandler        StreamErrorHandlerFunc
	routingErrorHandler       RoutingErrorHandlerFunc
	disablePathLengthFallback bool
}

func NewServeMux(opts ...ServeMuxOption) *ServeMux {
	return &ServeMux{}
}

func (s *ServeMux) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
}

func (s *ServeMux) Router() *httprouter.Router {
	return s.router
}

// MarshalerForRequest returns the inbound/outbound marshalers for this request.
// It checks the registry on the ServeMux for the MIME type set by the Content-Type header.
// If it isn't set (or the request Content-Type is empty), checks for "*".
// If there are multiple Content-Type headers set, choose the first one that it can
// exactly match in the registry.
// Otherwise, it follows the above logic for "*"/InboundMarshaler/OutboundMarshaler.
func (s *ServeMux) MarshalerForRequest(req *http.Request) (inbound, outbound Marshaler) {
	for _, acceptVal := range req.Header[marshal.AcceptHeader] {
		if m, ok := s.marshalers.MIMEMap[acceptVal]; ok {
			outbound = m
			break
		}
	}

	for _, contentTypeVal := range req.Header[marshal.ContentTypeHeader] {
		contentType, _, err := mime.ParseMediaType(contentTypeVal)
		if err != nil {
			grpclog.Infof("Failed to parse Content-Type %s: %v", contentTypeVal, err)
			continue
		}
		if m, ok := s.marshalers.MIMEMap[contentType]; ok {
			inbound = m
			break
		}
	}

	if inbound == nil {
		inbound = s.marshalers.MIMEMap[marshal.MIMEWildcard]
	}
	if outbound == nil {
		outbound = inbound
	}

	return inbound, outbound
}
