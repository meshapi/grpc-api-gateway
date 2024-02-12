package gateway

import (
	"context"
	"net/http"
	"net/textproto"
	"strings"

	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/proto"
)

func (s *ServeMux) handleForwardResponseServerMetadata(w http.ResponseWriter, md ServerMetadata) {
	for k, vs := range md.HeaderMD {
		if h, ok := s.outgoingHeaderMatcher(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}
}

func (s *ServeMux) handleForwardResponseOptions(ctx context.Context, w http.ResponseWriter, receivedResponse proto.Message) error {
	if len(s.forwardResponseOptions) == 0 {
		return nil
	}

	for _, opt := range s.forwardResponseOptions {
		if err := opt(ctx, w, receivedResponse); err != nil {
			grpclog.Infof("Error handling ForwardResponseOptions: %v", err)
			return err
		}
	}

	return nil
}

func (s *ServeMux) handleForwardResponseStreamErrorChunked(
	ctx context.Context,
	wroteHeader bool,
	marshaler Marshaler,
	writer http.ResponseWriter,
	req *http.Request,
	err error,
	delimiter []byte) {

	status, msg := s.streamErrorHandler(ctx, req, err)
	if !wroteHeader {
		writer.Header().Set("Content-Type", marshaler.ContentType(msg))
		writer.WriteHeader(status)
	}

	if msg != nil {
		buf, err := marshaler.Marshal(msg)
		if err != nil {
			grpclog.Infof("Failed to marshal an error: %v", err)
			return
		}
		if _, err := writer.Write(buf); err != nil {
			grpclog.Infof("Failed to notify error to client: %v", err)
			return
		}
	}

	if _, err := writer.Write(delimiter); err != nil {
		grpclog.Infof("Failed to send delimiter chunk: %v", err)
		return
	}
}

func handleForwardResponseTrailer(w http.ResponseWriter, md ServerMetadata) {
	for k, vs := range md.TrailerMD {
		tKey := MetadataTrailerPrefix + k
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}

func handleForwardResponseTrailerHeader(w http.ResponseWriter, md ServerMetadata) {
	for k := range md.TrailerMD {
		tKey := textproto.CanonicalMIMEHeaderKey(MetadataTrailerPrefix + k)
		w.Header().Add("Trailer", tKey)
	}
}

func requestAcceptsTrailers(req *http.Request) bool {
	te := req.Header.Get("TE")
	return strings.Contains(strings.ToLower(te), "trailers")
}
