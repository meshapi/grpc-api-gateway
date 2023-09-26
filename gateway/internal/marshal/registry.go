package marshal

import (
	"errors"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
)

// MIMEWildcard is the fallback MIME type used for requests which do not match
// a registered MIME type.
const MIMEWildcard = "*"

var (
	AcceptHeader      = http.CanonicalHeaderKey("Accept")
	ContentTypeHeader = http.CanonicalHeaderKey("Content-Type")

	DefaultMarshaler = &HTTPBodyMarshaler{
		Marshaler: &JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
	}
)

// Registry is a mapping from MIME types to Marshalers.
type Registry struct {
	MIMEMap map[string]Marshaler
}

// Add adds a marshaler for a case-sensitive MIME type string ("*" to match any
// MIME type).
func (r Registry) Add(mime string, marshaler Marshaler) error {
	if len(mime) == 0 {
		return errors.New("empty MIME type")
	}

	r.MIMEMap[mime] = marshaler

	return nil
}

// NewMarshalerMIMERegistry returns a new registry of marshalers.
// It allows for a mapping of case-sensitive Content-Type MIME type string to runtime.Marshaler interfaces.
//
// For example, you could allow the client to specify the use of the runtime.JSONPb marshaler
// with a "application/jsonpb" Content-Type and the use of the runtime.JSONBuiltin marshaler
// with a "application/json" Content-Type.
// "*" can be used to match any Content-Type.
// This can be attached to a ServerMux with the marshaler option.
func NewMarshalerMIMERegistry() Registry {
	return Registry{
		MIMEMap: map[string]Marshaler{
			MIMEWildcard: DefaultMarshaler,
		},
	}
}
