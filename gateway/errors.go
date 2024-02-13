package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/meshapi/grpc-rest-gateway/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

type (
	// ErrRouting is an error type specific to routing errors such as not found and method not allowed.
	ErrRouting uint8

	// ErrMarshal is the error returned by the gateway when marshal or unmarshal functions fail.
	ErrMarshal struct {
		// Err is the underlying marshal/unmarshal error.
		Err error
		// Inbound indicates whether or not the error is from unmarshaling an inbound request payload.
		//
		// If this value is true, it indicates that the request payload could not be unmarshaled
		// into the expected gRPC request type.
		//
		// If this value is false, it indicates the marshaler failed to encode response payload when
		// forwarding the response back to the HTTP client.
		Inbound bool
	}

	// ErrPathParameterMissing is the error generated by the gateway when HTTP path parameters are missing.
	ErrPathParameterMissing struct {
		// Name is the name of the parameter that happens to be a field path in the proto type.
		Name string
	}

	// ErrPathParameterTypeMismatch is the error generated by the gateway when path parameter type is unexpected.
	ErrPathParameterTypeMismatch struct {
		// Err is the underlying type mismatch.
		Err error
		// Name is the name of the parameter that happens to be a field path in the proto type.
		Name string
	}

	// ErrPathParameterInvalidEnum is the error generated when parsing values path parameters to an enum type fails.
	ErrPathParameterInvalidEnum struct {
		// Err is the underlying parsing error.
		Err error
		// Name is the name of the parameter that happens to be a field path in the proto type.
		Name string
	}

	// ErrInvalidQueryParameters is the error related to parsing query params. If the query parameters are invalid, cannot be
	// parsed or their values cannot be parsed, this error type gets used.
	ErrInvalidQueryParameters struct {
		// Err is the underlying parsing error.
		Err error
	}

	// ErrStreamingMethodNotAllowed is the error related to when a server receives a request that does not use any of the
	// available streaming methods.
	ErrStreamingMethodNotAllowed struct {
		// MethodSupportsWebsocket indicates whether or not the server accepts websocket for this method.
		MethodSupportsWebsocket bool
		// MethodSupportsSSE indicates whether or not the server accepts SSE for this method.
		MethodSupportsSSE bool
		// MethodSupportsChunkedTransfer indicates whether or not the server accepts chunked transfer streaming.
		MethodSupportsChunkedTransfer bool
	}
)

const (
	// ErrRoutingMethodNotAllowed is for routes that exist but with a different HTTP method.
	ErrRoutingMethodNotAllowed ErrRouting = iota
	// ErrRoutingNotFound is for routes that are not handled by the serve mux.
	ErrRoutingNotFound
)

func (r ErrRouting) Error() string {
	switch r {
	case ErrRoutingMethodNotAllowed:
		return "Method Not Allowed"
	case ErrRoutingNotFound:
		return "Not Found"
	default:
		return "Internal Server Error"
	}
}

func (r ErrRouting) GRPCStatus() *status.Status {
	switch r {
	case ErrRoutingMethodNotAllowed:
		return status.New(codes.Unimplemented, "Method Not Allowed")
	case ErrRoutingNotFound:
		return status.New(codes.NotFound, "Not Found")
	default:
		return status.New(codes.Internal, "Internal Server Error")
	}
}

func (e ErrMarshal) Error() string {
	return e.Err.Error()
}

func (e ErrMarshal) GRPCStatus() *status.Status {
	if e.Inbound {
		return status.New(codes.InvalidArgument, e.Error())
	}

	return status.New(codes.Internal, e.Error())
}

func (e ErrPathParameterMissing) Error() string {
	return fmt.Sprintf("missing path parameter: %v", e.Name)
}

func (e ErrPathParameterMissing) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

func (e ErrPathParameterTypeMismatch) Error() string {
	return fmt.Sprintf("type mismatch, path parameter: %q, error: %v", e.Name, e.Err)
}

func (e ErrPathParameterTypeMismatch) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

func (e ErrPathParameterInvalidEnum) Error() string {
	return fmt.Sprintf("could not parse path as enum value, path parameter: %q, error: %v", e.Name, e.Err)
}

func (e ErrPathParameterInvalidEnum) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

func (e ErrInvalidQueryParameters) Error() string {
	return e.Err.Error()
}

func (e ErrInvalidQueryParameters) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

func (e ErrStreamingMethodNotAllowed) Error() string {
	return "Streaming Method Not Allowed"
}

func (e ErrStreamingMethodNotAllowed) GRPCStatus() *status.Status {
	return status.New(codes.Unimplemented, e.Error())
}

type (
	// ErrorHandlerFunc is the signature used to configure error handling.
	ErrorHandlerFunc func(context.Context, *ServeMux, Marshaler, http.ResponseWriter, *http.Request, error)

	// StreamErrorHandlerFunc is the signature used to configure stream error handling.
	//
	// This function should return desired HTTP status code for the response and any object.
	// The status would be used only if status has not already been written.
	// The data is flexible format and will be unmarshaled using the marshaler for the request.
	// If the returning data is nil, no data will be written.
	//
	// This is used for chunked transfer only.
	StreamErrorHandlerFunc func(context.Context, *http.Request, error) (httpStatus int, data any)

	// SSEErrorHandlerFunc is the signature used to configure SSE error handling.
	//
	// This function should return desired SSE message for the error response. If the returned message is nil, no
	// response is sent.
	SSEErrorHandlerFunc func(context.Context, Marshaler, *http.Request, error) *SSEMessage

	// WebsocketErrorHandlerFunc is the signature used to configure websocket error handling.
	//
	// This function should return desired error message that will get marshaled and sent back to the client before
	// closing the connection.
	//
	// If the returned data is nil, no message is sent back before closing the connection.
	WebsocketErrorHandlerFunc func(context.Context, Marshaler, *http.Request, websocket.Connection, error)

	// RoutingErrorHandlerFunc is the signature used to configure error handling for routing errors.
	RoutingErrorHandlerFunc func(context.Context, *ServeMux, Marshaler, http.ResponseWriter, *http.Request, ErrRouting)
)

// HTTPStatusError is the error to use when needing to provide a different HTTP status code for an error
// passed to the DefaultRoutingErrorHandler.
type HTTPStatusError struct {
	HTTPStatus int
	Err        error
}

func (e HTTPStatusError) Error() string {
	return e.Err.Error()
}

// HTTPStatusFromCode converts a gRPC error code into the corresponding HTTP response status.
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func HTTPStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// Note, this deliberately doesn't translate to the similarly named '412 Precondition Failed' HTTP response status.
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		grpclog.Infof("Unknown gRPC error code: %v", code)
		return http.StatusInternalServerError
	}
}

// HTTPError uses the mux-configured error handler.
func (s *ServeMux) HTTPError(
	ctx context.Context, marshaler Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	s.errorHandler(ctx, s, marshaler, w, r, err)
}

// WebsocketError uses the mux-configured websocket error handler.
func (s *ServeMux) WebsocketError(
	ctx context.Context, marshaler Marshaler, r *http.Request, c websocket.Connection, err error) {
	s.websocketErrorHandler(ctx, marshaler, r, c, err)
}

// DefaultHTTPErrorHandler is the default error handler.
// If "err" is a gRPC Status, the function replies with the status code mapped by HTTPStatusFromCode.
// If "err" is a HTTPStatusError, the function replies with the status code provide by that struct. This is
// intended to allow passing through of specific statuses via the function set via WithRoutingErrorHandler
// for the ServeMux constructor to handle edge cases which the standard mappings in HTTPStatusFromCode
// are insufficient for.
// If otherwise, it replies with http.StatusInternalServerError.
//
// The response body written by this function is a Status message marshaled by the Marshaler.
func DefaultHTTPErrorHandler(ctx context.Context, mux *ServeMux, marshaler Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	// return Internal when Marshal failed
	const fallback = `{"code": 13, "message": "failed to marshal error message"}`

	var customStatus HTTPStatusError
	if errors.As(err, &customStatus) {
		err = customStatus.Err
	}

	s := status.Convert(err)
	pb := s.Proto()

	w.Header().Del("Trailer")
	w.Header().Del("Transfer-Encoding")

	contentType := marshaler.ContentType(pb)
	w.Header().Set("Content-Type", contentType)

	if s.Code() == codes.Unauthenticated {
		w.Header().Set("WWW-Authenticate", s.Message())
	}

	buf, merr := marshaler.Marshal(pb)
	if merr != nil {
		grpclog.Infof("Failed to marshal error message %q: %v", s, merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			grpclog.Infof("Failed to write response: %v", err)
		}
		return
	}

	md, ok := ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("Failed to extract ServerMetadata from context")
	}

	mux.handleForwardResponseServerMetadata(w, md)

	// RFC 7230 https://tools.ietf.org/html/rfc7230#section-4.1.2
	// Unless the request includes a TE header field indicating "trailers"
	// is acceptable, as described in Section 4.3, a server SHOULD NOT
	// generate trailer fields that it believes are necessary for the user
	// agent to receive.
	doForwardTrailers := requestAcceptsTrailers(r)

	if doForwardTrailers {
		handleForwardResponseTrailerHeader(w, md)
		w.Header().Set("Transfer-Encoding", "chunked")
	}

	st := HTTPStatusFromCode(s.Code())
	if customStatus.HTTPStatus >= 100 && customStatus.HTTPStatus < 600 {
		st = customStatus.HTTPStatus
	}

	w.WriteHeader(st)
	if _, err := w.Write(buf); err != nil {
		grpclog.Infof("Failed to write response: %v", err)
	}

	if doForwardTrailers {
		handleForwardResponseTrailer(w, md)
	}
}

func DefaultStreamErrorHandler(_ context.Context, _ *http.Request, err error) (int, any) {
	st := status.Convert(err)
	return HTTPStatusFromCode(st.Code()), st.Proto()
}

func DefaultSSEErrorHandler(_ context.Context, marshaler Marshaler, _ *http.Request, err error) *SSEMessage {
	const fallback = `{"code": 13, "message": "failed to marshal error message"}`

	message := &SSEMessage{
		Event: "failure",
	}

	st := status.Convert(err)
	message.Data, err = marshaler.Marshal(st.Proto())
	if err != nil {
		grpclog.Infof("Failed to marshal SSE error message: %v", err)
		message.Data = []byte(fallback)
	}

	return message
}

// DefaultRoutingErrorHandler is our default handler for routing errors.
// By default http error codes mapped on the following error codes:
//
//	NotFound -> grpc.NotFound
//	StatusBadRequest -> grpc.InvalidArgument
//	MethodNotAllowed -> grpc.Unimplemented
//	Other -> grpc.Internal, method is not expecting to be called for anything else
func DefaultRoutingErrorHandler(
	ctx context.Context, mux *ServeMux, marshaler Marshaler, w http.ResponseWriter, r *http.Request, err ErrRouting) {

	statusCode := http.StatusInternalServerError
	switch err {
	case ErrRoutingMethodNotAllowed:
		statusCode = http.StatusMethodNotAllowed
	case ErrRoutingNotFound:
		statusCode = http.StatusNotFound
	}

	mux.errorHandler(ctx, mux, marshaler, w, r, HTTPStatusError{HTTPStatus: statusCode, Err: err})
}

func DefaultWebsocketErrorHandler(
	ctx context.Context, marshaler Marshaler, _ *http.Request, connection websocket.Connection, err error) {
	defer connection.Close()
	message := status.Convert(err)
	data, uerr := marshaler.Marshal(message.Proto())
	if uerr != nil {
		grpclog.Infof("failed to marshal websocket error: %s", uerr)
		return
	}
	if err := connection.SendMessage(data); err != nil && err != io.EOF {
		grpclog.Infof("failed to send websocket error: %s", err)
	}
}
