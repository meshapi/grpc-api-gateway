package internal

// ProtoTypeName describes a proto name and the number of outer segments to keep context of the count of parents for
// nested types.
type ProtoTypeName struct {
	// FQN is the fully qualified name of the proto type.
	FQN string
	// OuterLength is the number of parent message types.
	OuterLength uint8
}
