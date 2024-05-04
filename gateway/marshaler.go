package gateway

import "github.com/meshapi/grpc-api-gateway/protomarshal"

// Marshaler defines a conversion between byte sequence and gRPC payloads / fields.
type Marshaler = protomarshal.Marshaler

// Decoder decodes a byte sequence
type Decoder = protomarshal.Decoder

// Encoder encodes gRPC payloads / fields into byte sequence.
type Encoder = protomarshal.Encoder

// DecoderFunc adapts an decoder function into Decoder.
type DecoderFunc = protomarshal.DecoderFunc

// EncoderFunc adapts an encoder function into Encoder
type EncoderFunc = protomarshal.EncoderFunc

// Delimited defines the streaming delimiter.
type Delimited = protomarshal.Delimited
