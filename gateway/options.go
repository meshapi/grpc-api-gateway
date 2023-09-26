package gateway

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// ServeMuxOption is an option that can be given to a ServeMux on construction.
type ServeMuxOption interface {
	apply(*ServeMux)
}

type optionFunc func(*ServeMux)

func (o optionFunc) apply(c *ServeMux) {
	o(c)
}

// WithForwardResponseOption returns a ServeMuxOption representing the forwardResponseOption.
//
// forwardResponseOption is an option that will be called on the relevant context.Context,
// http.ResponseWriter, and proto.Message before every forwarded response.
//
// The message may be nil in the case where just a header is being sent.
func WithForwardResponseOption(forwardResponseOption ForwardResponseFunc) ServeMuxOption {
	return optionFunc(func(s *ServeMux) {
		s.forwardResponseOptions = append(s.forwardResponseOptions, forwardResponseOption)
	})
}

// SetQueryParameterParser sets the query parameter parser, used to populate message from query parameters.
// Configuring this will mean the generated OpenAPI output is no longer correct, and it should be
// done with careful consideration.
func SetQueryParameterParser(queryParameterParser QueryParameterParser) ServeMuxOption {
	return optionFunc(func(s *ServeMux) {
		s.queryParamParser = queryParameterParser
	})
}

// WithIncomingHeaderMatcher returns a ServeMuxOption representing a headerMatcher for incoming request to gateway.
//
// This matcher will be called with each header in http.Request. If matcher returns true, that header will be
// passed to gRPC context. To transform the header before passing to gRPC context, matcher should return modified header.
func WithIncomingHeaderMatcher(fn HeaderMatcherFunc) ServeMuxOption {
	for _, header := range fn.matchedMalformedHeaders() {
		grpclog.Warningf("The configured forwarding filter would allow %q to be sent to the gRPC server, which will likely cause errors. See https://github.com/grpc/grpc-go/pull/4803#issuecomment-986093310 for more information.", header)
	}

	return optionFunc(func(s *ServeMux) {
		s.incomingHeaderMatcher = fn
	})
}

// matchedMalformedHeaders returns the malformed headers that would be forwarded to gRPC server.
func (fn HeaderMatcherFunc) matchedMalformedHeaders() []string {
	if fn == nil {
		return nil
	}
	headers := make([]string, 0)
	for header := range malformedHTTPHeaders {
		out, accept := fn(header)
		if accept && isMalformedHTTPHeader(out) {
			headers = append(headers, out)
		}
	}
	return headers
}

// WithOutgoingHeaderMatcher returns a ServeMuxOption representing a headerMatcher for outgoing response from gateway.
//
// This matcher will be called with each header in response header metadata. If matcher returns true, that header will be
// passed to http response returned from gateway. To transform the header before passing to response,
// matcher should return modified header.
func WithOutgoingHeaderMatcher(fn HeaderMatcherFunc) ServeMuxOption {
	return optionFunc(func(s *ServeMux) {
		s.outgoingHeaderMatcher = fn
	})
}

// WithMetadata returns a ServeMuxOption for passing metadata to a gRPC context.
//
// This can be used by services that need to read from http.Request and modify gRPC context. A common use case
// is reading token from cookie and adding it in gRPC context.
func WithMetadata(annotator MetadataAnnotatorFunc) ServeMuxOption {
	return optionFunc(func(s *ServeMux) {
		s.metadataAnnotators = append(s.metadataAnnotators, annotator)
	})
}

// WithErrorHandler returns a ServeMuxOption for configuring a custom error handler.
//
// This can be used to configure a custom error response.
func WithErrorHandler(fn ErrorHandlerFunc) ServeMuxOption {
	return optionFunc(func(s *ServeMux) {
		s.errorHandler = fn
	})
}

// WithStreamErrorHandler returns a ServeMuxOption that will use the given custom stream
// error handler, which allows for customizing the error trailer for server-streaming
// calls.
//
// For stream errors that occur before any response has been written, the mux's
// ErrorHandler will be invoked. However, once data has been written, the errors must
// be handled differently: they must be included in the response body. The response body's
// final message will include the error details returned by the stream error handler.
func WithStreamErrorHandler(fn StreamErrorHandlerFunc) ServeMuxOption {
	return optionFunc(func(s *ServeMux) {
		s.streamErrorHandler = fn
	})
}

// WithRoutingErrorHandler returns a ServeMuxOption for configuring a custom error handler to  handle http routing errors.
//
// Method called for errors which can happen before gRPC route selected or executed.
// The following error codes: StatusMethodNotAllowed StatusNotFound StatusBadRequest
func WithRoutingErrorHandler(fn RoutingErrorHandlerFunc) ServeMuxOption {
	return optionFunc(func(s *ServeMux) {
		s.routingErrorHandler = fn
	})
}

// WithDisablePathLengthFallback returns a ServeMuxOption for disable path length fallback.
func WithDisablePathLengthFallback() ServeMuxOption {
	return optionFunc(func(s *ServeMux) {
		s.disablePathLengthFallback = true
	})
}

// WithHealthEndpointAt returns a ServeMuxOption that will add an endpoint to the created ServeMux at the path specified by endpointPath.
// When called the handler will forward the request to the upstream grpc service health check (defined in the
// gRPC Health Checking Protocol).
//
// See here https://grpc-ecosystem.github.io/grpc-gateway/docs/operations/health_check/ for more information on how
// to setup the protocol in the grpc server.
//
// If you define a service as query parameter, this will also be forwarded as service in the HealthCheckRequest.
func WithHealthEndpointAt(healthCheckClient grpc_health_v1.HealthClient, endpointPath string) ServeMuxOption {
	return optionFunc(func(s *ServeMux) {
		// error can be ignored since pattern is definitely valid
		s.router.GET(endpointPath, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			_, outboundMarshaler := s.MarshalerForRequest(r)

			resp, err := healthCheckClient.Check(r.Context(), &grpc_health_v1.HealthCheckRequest{
				Service: r.URL.Query().Get("service"),
			})
			if err != nil {
				s.errorHandler(r.Context(), s, outboundMarshaler, w, r, err)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
				switch resp.GetStatus() {
				case grpc_health_v1.HealthCheckResponse_NOT_SERVING, grpc_health_v1.HealthCheckResponse_UNKNOWN:
					err = status.Error(codes.Unavailable, resp.String())
				case grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN:
					err = status.Error(codes.NotFound, resp.String())
				}

				s.errorHandler(r.Context(), s, outboundMarshaler, w, r, err)
				return
			}

			_ = outboundMarshaler.NewEncoder(w).Encode(resp)

		})
	})
}

// WithHealthzEndpoint returns a ServeMuxOption that will add a /healthz endpoint to the created ServeMux.
//
// See WithHealthEndpointAt for the general implementation.
func WithHealthzEndpoint(healthCheckClient grpc_health_v1.HealthClient) ServeMuxOption {
	return WithHealthEndpointAt(healthCheckClient, "/healthz")
}
