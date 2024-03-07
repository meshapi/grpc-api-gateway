package openapiv3

import (
	"fmt"
)

type Extensions map[string]any

// Document is the OpenAPI top-level document.
type Document struct {
	OpenAPI               string                 `json:"openapi" yaml:"openapi"`
	Info                  *Info                  `json:"info" yaml:"info" validate:"required"`
	SchemaDialect         string                 `json:"jsonSchemaDialect,omitempty" yaml:"jsonSchemaDialect,omitempty"`
	Servers               []Server               `json:"servers,omitempty" yaml:"servers,omitempty"`
	Security              map[string][]string    `json:"security,omitempty" yaml:"security,omitempty"`
	Tags                  []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocumentation *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Extensions            Extensions             `json:"-" yaml:"-"`
}

func (d *Document) Validate() error {
	if d.OpenAPI != "3.1" {
		return fmt.Errorf("expected OpenAPI v3.1, instead got: %q", d.OpenAPI)
	}

	return Validate(d, d)
}
