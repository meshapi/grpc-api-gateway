package openapiv3

type ComponentsCore struct {
	Schemas   map[string]*Extensible[SchemaCore] `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses map[string]*Ref[Response]          `json:"responses,omitempty" yaml:"responses,omitempty"`
}

type ExampleCore struct {
	Summary       string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string `json:"description,omitempty" yaml:"description,omitempty"`
	Value         any    `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

type HeaderCore struct {
	Description     string                                   `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                                     `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                                     `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                                     `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string                                   `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool                                     `json:"explode,omitempty" yaml:"explode,omitempty"`
	Schema          *Extensible[SchemaCore]                  `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         any                                      `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]*Ref[Extensible[ExampleCore]] `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content         map[string]*Extensible[MediaTypeCore]    `json:"content,omitempty" yaml:"content,omitempty"`
}

type EncodingCore struct {
	ContentType   string                  `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]*Ref[Header] `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string                  `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool                    `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool                    `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

type MediaTypeCore struct {
	Schema   *Extensible[SchemaCore]                  `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  any                                      `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]*Ref[Extensible[ExampleCore]] `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]*Extensible[EncodingCore]     `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

type LinkCore struct {
	OperationID  string         `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	OperationRef string         `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  any            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string         `json:"description,omitempty" yaml:"description,omitempty"`
	Server       *Server        `json:"server,omitempty" yaml:"server,omitempty"`
}

type ResponseCore struct {
	Description string                                  `json:"description,omitempty" yaml:"description,omitempty"`
	Headers     map[string]*Ref[Extensible[HeaderCore]] `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]*Extensible[MediaTypeCore]   `json:"content,omitempty" yaml:"content,omitempty"`
	Links       map[string]*Ref[Extensible[LinkCore]]   `json:"links,omitempty" yaml:"links,omitempty"`
}
