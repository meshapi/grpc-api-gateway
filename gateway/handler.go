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
