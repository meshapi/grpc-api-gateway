package openapiv3

import (
	"encoding/json"
)

const (
	TypeObject  = "object"
	TypeString  = "string"
	TypeArray   = "array"
	TypeNumber  = "number"
	TypeInteger = "integer"
	TypeBoolean = "boolean"
	TypeNull    = "null"
)

type DiscriminatorCore struct {
	PropertyName string            `json:"propertyName,omitempty" yaml:"propertyName,omitempty"`
	Mapping      map[string]string `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

type SchemaCore struct {
	Discriminator         *Discriminator         `json:"discriminator,omitempty" yaml:"discriminator,omitempty"`
	ExternalDocumentation *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Ref                   string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Schema                string                 `json:"$schema,omitempty" yaml:"$schema,omitempty"`

	Title                string             `json:"title,omitempty" yaml:"title,omitempty"`
	Pattern              string             `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Required             []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                 []string           `json:"enum,omitempty" yaml:"enum,omitempty"`
	MultipleOf           float64            `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum              float64            `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ExclusiveMaximum     float64            `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	Minimum              float64            `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMinimum     float64            `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	MaxLength            uint64             `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength            uint64             `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxItems             uint64             `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems             uint64             `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems          bool               `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties        uint64             `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        uint64             `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Type                 TypeSet            `json:"type,omitempty" yaml:"type,omitempty"`
	Description          string             `json:"description,omitempty" yaml:"description,omitempty"`
	Items                *ItemSpec          `json:"items,omitempty" yaml:"items,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Default              any                `json:"default,omitempty" yaml:"default,omitempty"`
	AllOf                []*Schema          `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	AnyOf                []*Schema          `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	OneOf                []*Schema          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	Not                  *Schema            `json:"not,omitempty" yaml:"not,omitempty"`
	ReadOnly             bool               `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool               `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Examples             []any              `json:"examples,omitempty" yaml:"examples,omitempty"`
	Format               string             `json:"format,omitempty" yaml:"format,omitempty"`
}

type TypeSet []string

func (t TypeSet) MarshalJSON() ([]byte, error) {
	if len(t) == 1 {
		return json.Marshal(t[0])
	}

	return json.Marshal([]string(t))
}

func (t TypeSet) MarshalYAML() (any, error) {
	if len(t) == 1 {
		return t[0], nil
	}

	return []string(t), nil
}

// ItemSpec is used to generate correct "items" value in JSON schema.
// If Schema is defined, schema gets marshaled, otherwise, items list gets used.
type ItemSpec struct {
	Schema *Schema
	Items  []*Schema
}

func (i *ItemSpec) MarshalJSON() ([]byte, error) {
	if i.Schema != nil {
		return json.Marshal(i.Schema)
	}

	return json.Marshal(i.Items)
}

func (i *ItemSpec) MarshalYAML() (any, error) {
	if i.Schema != nil {
		return i.Schema, nil
	}

	return i.Items, nil
}
