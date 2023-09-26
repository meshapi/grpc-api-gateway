package gateway

import "github.com/meshapi/grpc-rest-gateway/gateway/internal/marshal"

// Marshaler defines a conversion between byte sequence and gRPC payloads / fields.
type Marshaler = marshal.Marshaler

// Decoder decodes a byte sequence
type Decoder = marshal.Decoder

// Encoder encodes gRPC payloads / fields into byte sequence.
type Encoder = marshal.Encoder

// DecoderFunc adapts an decoder function into Decoder.
type DecoderFunc = marshal.DecoderFunc

// EncoderFunc adapts an encoder function into Encoder
type EncoderFunc = marshal.EncoderFunc

// Delimited defines the streaming delimiter.
type Delimited = marshal.Delimited
