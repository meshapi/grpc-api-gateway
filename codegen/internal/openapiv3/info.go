package openapiv3

type Contact struct {
	Name  string `json:"name" yaml:"name"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

type License struct {
	Name       string `json:"name" yaml:"name" validate:"required"`
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	URL        string `json:"url,omitempty" yaml:"url,omitempty"`
}

type Info struct {
	Title          string               `json:"title" yaml:"title" validate:"required"`
	Summary        string               `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description    string               `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string               `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Extensible[Contact] `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *Extensible[License] `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string               `json:"version" yaml:"version" validate:"required"`
}

type ServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default" validate:"required"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

type Server struct {
	URL         string                                `json:"url" yaml:"url" validate:"required"`
	Description string                                `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]Extensible[ServerVariable] `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type ExternalDocumentation struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url" validate:"required"`
}

type Tag struct {
	Name         string                             `json:"name" yaml:"name" validate:"required"`
	Description  string                             `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *Extensible[ExternalDocumentation] `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}
