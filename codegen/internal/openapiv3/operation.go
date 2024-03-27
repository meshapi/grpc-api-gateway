package openapiv3

type PathCore struct {
	Summary     string         `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *OperationCore `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *OperationCore `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *OperationCore `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *OperationCore `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *OperationCore `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *OperationCore `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *OperationCore `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *OperationCore `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []*Server      `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []*Ref[Server] `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

type OperationCore struct {
	Tags         []string                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary      string                    `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                    `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation    `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	OperationID  string                    `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   []*Ref[Parameter]         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *Ref[RequestBody]         `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses    map[string]*Ref[Response] `json:"responses,omitempty" yaml:"responses,omitempty"`
	Deprecated   bool                      `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security     []map[string][]string     `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      []*Server                 `json:"servers,omitempty" yaml:"servers,omitempty"`
}
