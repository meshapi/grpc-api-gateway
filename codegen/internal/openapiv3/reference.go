package openapiv3

import "encoding/json"

type Reference struct {
	Ref         string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Ref is a container that allows either using a reference or data.
type Ref[T any] struct {
	Data      T
	Reference *Reference
}

func (r Ref[T]) MarshalJSON() ([]byte, error) {
	if r.Reference != nil {
		return json.Marshal(r.Reference)
	}

	return json.Marshal(r.Data)
}

func (r Ref[T]) MarshalYAML() (any, error) {
	if r.Reference != nil {
		return r.Reference, nil
	}

	return r.Data, nil
}
