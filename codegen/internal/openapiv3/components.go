package openapiv3

type Components struct {
	Schemas map[string]*Extensible[Schema] `json:"schemas,omitempty" yaml:"schemas,omitempty"`
}
