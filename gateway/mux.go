package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/textproto"
	"strings"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/meshapi/grpc-rest-gateway/gateway/internal/marshal"
	"github.com/meshapi/grpc-rest-gateway/websocket"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// HeaderMatcherFunc checks whether a header key should be forwarded to/from gRPC context.
type HeaderMatcherFunc func(string) (string, bool)

// PanicHandlerFunc is a function that gets called when a panic is encountered.
type PanicHandlerFunc func(http.ResponseWriter, *http.Request, interface{})

// ForwardResponseFunc updates the outgoing gRPC request and the HTTP response.
type ForwardResponseFunc func(context.Context, http.ResponseWriter, proto.Message) error

// MetadataAnnotatorFunc updates the outgoing gRPC request context based on the incoming HTTP request.
type MetadataAnnotatorFunc func(context.Context, *http.Request) metadata.MD

// WebsocketUpgradeFunc handles a protocol upgrade and creates a websocket connection.
type WebsocketUpgradeFunc func(http.ResponseWriter, *http.Request) (websocket.Connection, error)

// DefaultHeaderMatcher is used to pass http request headers to/from gRPC context. This adds permanent HTTP header
// keys (as specified by the IANA, e.g: Accept, Cookie, Host) to the gRPC metadata with the grpcgateway- prefix. If you want to know which headers are considered permanent, you can view the isPermanentHTTPHeader function.
// HTTP headers that start with 'Grpc-Metadata-' are mapped to gRPC metadata after removing the prefix 'Grpc-Metadata-'.
// Other headers are not added to the gRPC metadata.
func DefaultHeaderMatcher(key string) (string, bool) {
	switch key = textproto.CanonicalMIMEHeaderKey(key); {
	case isPermanentHTTPHeader(key):
		return MetadataPrefix + key, true
	case strings.HasPrefix(key, MetadataHeaderPrefix):
		return key[len(MetadataHeaderPrefix):], true
	}
	return "", false
}

// ServeMux is a request multiplexer for grpc-gateway.
// It matches http requests to patterns and invokes the corresponding handler.
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
	websocketErrorHandler     WebsocketErrorHandlerFunc
	streamErrorHandler        StreamErrorHandlerFunc
	routingErrorHandler       RoutingErrorHandlerFunc
	websocketUpgradeFunc      WebsocketUpgradeFunc
	disablePathLengthFallback bool
}

// NewServeMux returns a new ServeMux whose internal mapping is empty.
func NewServeMux(opts ...ServeMuxOption) *ServeMux {
	mux := &ServeMux{
		router:                    httprouter.New(),
		queryParamParser:          &DefaultQueryParser{},
		forwardResponseOptions:    make([]ForwardResponseFunc, 0),
		marshalers:                marshal.NewMarshalerMIMERegistry(),
		metadataAnnotators:        nil,
		errorHandler:              DefaultHTTPErrorHandler,
		streamErrorHandler:        DefaultStreamErrorHandler,
		websocketErrorHandler:     DefaultWebsocketErrorHandler,
		routingErrorHandler:       DefaultRoutingErrorHandler,
		disablePathLengthFallback: false,
	}

	for _, opt := range opts {
		opt.apply(mux)
	}

	mux.router.HandleMethodNotAllowed = true
	mux.router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, outboundMarshaler := mux.MarshalerForRequest(r)
		mux.routingErrorHandler(r.Context(), mux, outboundMarshaler, w, r, ErrRoutingMethodNotAllowed)
	})

	mux.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, outboundMarshaler := mux.MarshalerForRequest(r)
		mux.routingErrorHandler(r.Context(), mux, outboundMarshaler, w, r, ErrRoutingNotFound)
	})

	if mux.incomingHeaderMatcher == nil {
		mux.incomingHeaderMatcher = DefaultHeaderMatcher
	}

	if mux.outgoingHeaderMatcher == nil {
		mux.outgoingHeaderMatcher = func(key string) (string, bool) {
			return fmt.Sprintf("%s%s", MetadataHeaderPrefix, key), true
		}
	}

	return mux
}

// UpgradeToWebsocket upgrades an HTTP request to a websocket connection.
func (s *ServeMux) UpgradeToWebsocket(response http.ResponseWriter, req *http.Request) (websocket.Connection, error) {
	if s.websocketUpgradeFunc != nil {
		return s.websocketUpgradeFunc(response, req)
	}

	return nil, errors.New("websockets are not supported in this server")
}

// ServeHTTP dispatches the request to the first handler whose pattern matches to r.Method and r.URL.Path.
func (s *ServeMux) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	// if explicitly requested and method overriding is enabled, change method.
	if override := req.Header.Get("X-HTTP-Method-Override"); override != "" && s.isPathLengthFallback(req) {
		req.Method = strings.ToUpper(override)
		if err := req.ParseForm(); err != nil {
			_, outboundMarshaler := s.MarshalerForRequest(req)
			sterr := status.Error(codes.InvalidArgument, err.Error())
			s.errorHandler(req.Context(), s, outboundMarshaler, writer, req, sterr)
			return
		}
	}

	// TODO: properly handle fallback to handle POST->GET. You can use LookUp here.
	s.router.ServeHTTP(writer, req)
}

// HandleWithParams registers a new handler for the method and pattern specified.
//
// NOTE: this method takes an httprouter.Handle function, helpful when path parameters are needed.
// if using http.Handler is desired, use Handle instead.
func (s *ServeMux) HandleWithParams(method, pattern string, handler httprouter.Handle) {
	s.router.Handle(method, pattern, handler)
}

// Handle registers a new handler for the method and pattern specified.
func (s *ServeMux) Handle(method, pattern string, handler http.Handler) {
	s.router.Handler(method, pattern, handler)
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

// IsWebsocketUpgrade returns whether or not the client is requesting for connection upgrade to websocket.
func (s *ServeMux) IsWebsocketUpgrade(req *http.Request) bool {
	if value := req.Header.Get("Connection"); value != "" && strings.ToLower(value) == "upgrade" {
		if upgrade := req.Header.Get("Upgrade"); upgrade != "" && strings.ToLower(upgrade) == "websocket" {
			return true
		}
	}
	return false
}

func (s *ServeMux) isPathLengthFallback(r *http.Request) bool {
	return !s.disablePathLengthFallback && r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/x-www-form-urlencoded"
}

func (s *ServeMux) ForwardResponseMessage(
	ctx context.Context,
	marshaler Marshaler,
	writer http.ResponseWriter,
	req *http.Request,
	receivedResponse proto.Message) {

	md, ok := ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("failed to extract ServerMetadata from context")
	}

	s.handleForwardResponseServerMetadata(writer, md)

	// RFC 7230 https://tools.ietf.org/html/rfc7230#section-4.1.2
	// Unless the request includes a TE header field indicating "trailers"
	// is acceptable, as described in Section 4.3, a server SHOULD NOT
	// generate trailer fields that it believes are necessary for the user
	// agent to receive.
	doForwardTrailers := requestAcceptsTrailers(req)

	if doForwardTrailers {
		handleForwardResponseTrailerHeader(writer, md)
		writer.Header().Set("Transfer-Encoding", "chunked")
	}

	contentType := marshaler.ContentType(receivedResponse)
	writer.Header().Set("Content-Type", contentType)

	if err := s.handleForwardResponseOptions(ctx, writer, receivedResponse); err != nil {
		s.HTTPError(ctx, marshaler, writer, req, err)
		return
	}

	var buf []byte
	var err error
	// TODO: find out how to select response keys.
	// TODO: can we use NewEncoder here to avoid the memory allocation here?
	buf, err = marshaler.Marshal(receivedResponse)
	if err != nil {
		grpclog.Infof("Marshal error: %v", err)
		s.HTTPError(ctx, marshaler, writer, req, ErrMarshal{Err: err, Inbound: false})
		return
	}

	if _, err = writer.Write(buf); err != nil {
		grpclog.Infof("Failed to write response: %v", err)
	}

	if doForwardTrailers {
		handleForwardResponseTrailer(writer, md)
	}
}

// ForwardResponseStreamChunked forwards the stream from gRPC server to REST client using Transfer-Encoding chunked.
func (s *ServeMux) ForwardResponseStreamChunked(
	ctx context.Context,
	marshaler Marshaler,
	writer http.ResponseWriter,
	req *http.Request,
	recv func() (proto.Message, error)) {

	f, ok := writer.(http.Flusher)
	if !ok {
		grpclog.Errorf("Flush not supported in %T", writer)
		http.Error(writer, "unexpected type of web server", http.StatusInternalServerError)
		return
	}

	md, ok := ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("Failed to extract ServerMetadata from context")
		http.Error(writer, "unexpected error", http.StatusInternalServerError)
		return
	}
	s.handleForwardResponseServerMetadata(writer, md)

	writer.Header().Set("Transfer-Encoding", "chunked")
	if err := s.handleForwardResponseOptions(ctx, writer, nil); err != nil {
		// TODO: Improve error handling here.
		s.HTTPError(ctx, marshaler, writer, req, err)
		return
	}

	var delimiter []byte
	if d, ok := marshaler.(Delimited); ok {
		delimiter = d.Delimiter()
	} else {
		delimiter = []byte("\n")
	}

	var wroteHeader bool
	for {
		resp, err := recv()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			s.handleForwardResponseStreamErrorChunked(ctx, wroteHeader, marshaler, writer, req, err, delimiter)
			return
		}
		if err := s.handleForwardResponseOptions(ctx, writer, resp); err != nil {
			s.handleForwardResponseStreamErrorChunked(ctx, wroteHeader, marshaler, writer, req, err, delimiter)
			return
		}

		if !wroteHeader {
			writer.Header().Set("Content-Type", marshaler.ContentType(resp))
		}

		var buf []byte
		// TODO: is this necessary when marshaler is a type that can handle http body?
		httpBody, isHTTPBody := resp.(*httpbody.HttpBody)
		switch {
		case resp == nil:
			buf, err = marshaler.Marshal(status.New(codes.Internal, "empty response"))
		case isHTTPBody:
			buf = httpBody.GetData()
		default:
			//result := map[string]interface{}{"result": resp}
			// TODO: find out how to select response keys.
			//if rb, ok := resp.(responseBody); ok {
			//  result["result"] = rb.XXX_ResponseBody()
			//}

			buf, err = marshaler.Marshal(resp)
		}

		if err != nil {
			grpclog.Infof("Failed to marshal response chunk: %v", err)
			s.handleForwardResponseStreamErrorChunked(ctx, wroteHeader, marshaler, writer, req, err, delimiter)
			return
		}
		if _, err := writer.Write(buf); err != nil {
			grpclog.Infof("Failed to send response chunk: %v", err)
			return
		}
		wroteHeader = true
		if _, err := writer.Write(delimiter); err != nil {
			grpclog.Infof("Failed to send delimiter chunk: %v", err)
			return
		}
		f.Flush()
	}
}

type ProtoMessage interface {
	proto.Message
	Reset()
}

func (s *ServeMux) ForwardWebsocket(
	ctx context.Context,
	req *http.Request, stream grpc.ClientStream, ws websocket.Connection,
	inboundMarshaler, outboundMarshaler Marshaler,
	protoReq, protoRes ProtoMessage) {

	closeWebsocketConnection := sync.OnceFunc(func() {
		if err := ws.Close(); err != nil {
			grpclog.Infof("Failed to close websocket connection: %v", err)
		}
	})
	defer closeWebsocketConnection()

	// receive from gRPC stream and forward to websocket.
	go func() {
		defer closeWebsocketConnection()

		for {
			protoRes.Reset()
			err := stream.RecvMsg(protoRes)
			if err == io.EOF {
				break
			}
			if err != nil {
				grpclog.Infof("Failed to receive message from gRPC stream: %v", err)
				break
			}
			data, err := outboundMarshaler.Marshal(protoRes)
			if err != nil {
				s.websocketErrorHandler(ctx, outboundMarshaler, req, ws, ErrMarshal{Err: err, Inbound: false})
				break
			}
			if err := ws.SendMessage(data); err != nil {
				grpclog.Infof("Failed to send gRPC response via websocket connection: %v", err)
				break
			}
		}
	}()

	// receive from websocket, forward to gRPC stream.
	for {
		data, err := ws.ReceiveMessage()
		if err == io.EOF {
			break
		}
		if err != nil {
			grpclog.Infof("failed to receive message: %v", err)
			break
		}
		protoReq.Reset()
		if err := inboundMarshaler.Unmarshal(data, protoReq); err != nil {
			grpclog.Infof("Failed to decode request from websocket: %v", err)
			s.websocketErrorHandler(ctx, outboundMarshaler, req, ws, ErrMarshal{Err: err, Inbound: true})
			break
		}

		if err := stream.SendMsg(protoReq); err != nil {
			if err != io.EOF {
				grpclog.Infof("Failed to send request from websocket: %v", err)
			}
			break
		}
	}
	if err := stream.CloseSend(); err != nil {
		grpclog.Infof("Failed to terminate gRPC client stream: %v")
	}
}

func (s *ServeMux) ForwardWebsocketServerStreaming(
	ctx context.Context,
	req *http.Request, stream grpc.ClientStream, ws websocket.Connection,
	outboundMarshaler Marshaler,
	protoRes ProtoMessage) {

	closeWebsocketConnection := sync.OnceFunc(func() {
		if err := ws.Close(); err != nil {
			grpclog.Infof("Failed to close websocket connection: %v", err)
		}
	})
	defer closeWebsocketConnection()

	// receive from gRPC stream and forward to websocket.
	go func() {
		defer closeWebsocketConnection()

		for {
			protoRes.Reset()
			err := stream.RecvMsg(protoRes)
			if err == io.EOF {
				break
			}
			if err != nil {
				grpclog.Infof("Failed to receive message from gRPC stream: %v", err)
				break
			}
			data, err := outboundMarshaler.Marshal(protoRes)
			if err != nil {
				s.websocketErrorHandler(ctx, outboundMarshaler, req, ws, ErrMarshal{Err: err, Inbound: false})
				break
			}
			if err := ws.SendMessage(data); err != nil {
				grpclog.Infof("Failed to send gRPC response via websocket connection: %v", err)
				break
			}
		}
	}()

	// receive from websocket, forward to gRPC stream.
	for {
		_, err := ws.ReceiveMessage()
		if err == io.EOF {
			break
		}
		if err != nil {
			grpclog.Infof("failed to receive message: %v", err)
			break
		}
	}
}
