package openapi

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
	Title          string   `json:"title" yaml:"title" validate:"required"`
	Summary        string   `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string   `json:"version" yaml:"version" validate:"required"`
}

type ServerVariable struct {
	Enum        []string `json:"enum" yaml:"enum"`
	Default     string   `json:"default" yaml:"default" validate:"required"`
	Description string   `json:"description" yaml:"description"`
}

type Server struct {
	URL         string                    `json:"url" yaml:"url"`
	Description string                    `json:"description" yaml:"description"`
	Variables   map[string]ServerVariable `json:"variables" yaml:"variables"`
}

type ExternalDocumentation struct {
	Description string `json:"description" yaml:"description"`
	URL         string `json:"url" yaml:"url" validate:"required"`
}

type Tag struct {
	Name         string                `json:"name" yaml:"name" validate:"required"`
	Description  string                `json:"description" yaml:"description"`
	ExternalDocs ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
}
