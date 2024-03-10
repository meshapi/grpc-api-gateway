package openapiv3

type ComponentsCore struct {
	Schemas         map[string]*Extensible[SchemaCore] `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       map[string]*Ref[Response]          `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]*Ref[Parameter]         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]*Ref[Example]           `json:"examples,omitempty" yaml:"examples,omitempty"`
	RequestBodies   map[string]*Ref[RequestBody]       `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         map[string]*Ref[Header]            `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]*Ref[SecurityScheme]    `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Links           map[string]*Ref[Link]              `json:"links,omitempty" yaml:"links,omitempty"`
}

type SecuritySchemeCore struct {
	Type             string      `json:"type,omitempty" yaml:"type,omitempty"`
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
}

type OAuthFlowsCore struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

type OAuthFlowCore struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

type RequestBodyCore struct {
	Description string                                `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]*Extensible[MediaTypeCore] `json:"content,omitempty" yaml:"content,omitempty"`
	Required    bool                                  `json:"required,omitempty" yaml:"required,omitempty"`
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

type ParameterCore struct {
	Name            string                                   `json:"name,omitempty" yaml:"name,omitempty"`
	In              string                                   `json:"in,omitempty" yaml:"in,omitempty"`
	Description     string                                   `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                                     `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                                     `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                                     `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	AllowReserved   bool                                     `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
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
