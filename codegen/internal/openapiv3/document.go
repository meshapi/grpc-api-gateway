package openapiv3

import (
	"fmt"
)

// The following are type aliases to simplify references to these types.
type (
	ExtendedDocument              = Extensible[Document]
	ExtendedInfo                  = Extensible[Info]
	ExtendedContact               = Extensible[Contact]
	ExtendedLicense               = Extensible[License]
	ExtendedServer                = Extensible[Server]
	ExtendedTag                   = Extensible[Tag]
	ExtendedExternalDocumentation = Extensible[ExternalDocumentation]
	ExtendedServerVariable        = Extensible[ServerVariable]
	ExtendedComponents            = Extensible[Components]
	ExtendedSchema                = Extensible[Schema]
)

const (
	Version = "3.1.0"
)

// Document is the OpenAPI top-level document.
type Document struct {
	OpenAPI       string               `json:"openapi" yaml:"openapi"`
	Info          *Extensible[Info]    `json:"info" yaml:"info" validate:"required"`
	SchemaDialect string               `json:"jsonSchemaDialect,omitempty" yaml:"jsonSchemaDialect,omitempty"`
	Servers       []Extensible[Server] `json:"servers,omitempty" yaml:"servers,omitempty"`
	// Paths
	Components            *Extensible[Components]            `json:"components,omitempty" yaml:"components"`
	Security              map[string][]string                `json:"security,omitempty" yaml:"security,omitempty"`
	Tags                  []Extensible[Tag]                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocumentation *Extensible[ExternalDocumentation] `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

func (d *Document) Validate() error {
	if d.OpenAPI != Version {
		return fmt.Errorf("expected OpenAPI %s, instead got: %q", Version, d.OpenAPI)
	}

	return Validate(d, d)
}
