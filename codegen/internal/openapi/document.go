package openapi

import "fmt"

// Document is the OpenAPI top-level document.
type Document struct {
	OpenAPI       string              `json:"openapi" yaml:"openapi"`
	Info          *Info               `json:"info" yaml:"info" validate:"required"`
	SchemaDialect string              `json:"jsonSchemaDialect" yaml:"jsonSchemaDialect"`
	Servers       []Server            `json:"servers" yaml:"servers"`
	Security      map[string][]string `json:"security" yaml:"security"`
	Tags          []Tag               `json:"tags" yaml:"tags"`
}

func (d *Document) Validate() error {
	if d.OpenAPI != "3.1" {
		return fmt.Errorf("expected OpenAPI v3.1, instead got: %q", d.OpenAPI)
	}

	return Validate(d, d)
}
