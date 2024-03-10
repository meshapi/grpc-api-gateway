package openapiv3

import (
	"fmt"
)

// The following are type aliases to simplify references to these types.
type (
	Document              = Extensible[DocumentCore]
	Info                  = Extensible[InfoCore]
	Contact               = Extensible[ContactCore]
	License               = Extensible[LicenseCore]
	Server                = Extensible[ServerCore]
	Tag                   = Extensible[TagCore]
	ExternalDocumentation = Extensible[ExternalDocumentationCore]
	ServerVariable        = Extensible[ServerVariableCore]
	Components            = Extensible[ComponentsCore]
	Schema                = Extensible[SchemaCore]
	Discriminator         = Extensible[DiscriminatorCore]
	Response              = Extensible[ResponseCore]
	Header                = Extensible[HeaderCore]
	Parameter             = Extensible[ParameterCore]
	MediaType             = Extensible[MediaTypeCore]
	Example               = Extensible[ExampleCore]
	Encoding              = Extensible[EncodingCore]
	Link                  = Extensible[LinkCore]
	RequestBody           = Extensible[RequestBodyCore]
)

const (
	Version = "3.1.0"
)

// DocumentCore is the OpenAPI top-level document.
type DocumentCore struct {
	OpenAPI       string    `json:"openapi" yaml:"openapi"`
	Info          *Info     `json:"info" yaml:"info" validate:"required"`
	SchemaDialect string    `json:"jsonSchemaDialect,omitempty" yaml:"jsonSchemaDialect,omitempty"`
	Servers       []*Server `json:"servers,omitempty" yaml:"servers,omitempty"`
	// Paths
	Components            *Components            `json:"components,omitempty" yaml:"components"`
	Security              map[string][]string    `json:"security,omitempty" yaml:"security,omitempty"`
	Tags                  []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocumentation *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

func (d *DocumentCore) Validate() error {
	if d.OpenAPI != Version {
		return fmt.Errorf("expected OpenAPI %s, instead got: %q", Version, d.OpenAPI)
	}

	return Validate(d, d)
}
