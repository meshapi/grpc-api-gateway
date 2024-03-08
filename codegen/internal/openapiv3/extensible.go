package openapiv3

import (
	"encoding/json"
	"reflect"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Extensions map[string]any

type Extensible[T any] struct {
	Object     T          `yaml:",inline"`
	Extensions Extensions `yaml:",inline"`
}

func (e Extensible[T]) MarshalJSON() ([]byte, error) {
	if len(e.Extensions) == 0 {
		return json.Marshal(e.Object)
	}
	return extensionMarshalJSON(e.Object, e.Extensions)
}

func extensionStructFieldName(key string) string {
	return strings.ReplaceAll(cases.Title(language.AmericanEnglish).String(key), "-", "_")
}

func extensionMarshalJSON(primaryObject any, extensions Extensions) ([]byte, error) {
	// To append arbitrary keys to the struct we'll render into json,
	// we're creating another struct that embeds the original one, and
	// its extra fields:
	//
	// The struct will look like
	// struct {
	//   *openapiCore
	//   XGrpcGatewayFoo json.RawMessage `json:"x-grpc-gateway-foo"`
	//   XGrpcGatewayBar json.RawMessage `json:"x-grpc-gateway-bar"`
	// }
	// and thus render into what we want -- the JSON of openapiCore with the
	// extensions appended.
	fields := []reflect.StructField{
		{ // embedded
			Name:      "Embedded",
			Type:      reflect.TypeOf(primaryObject),
			Anonymous: true,
		},
	}
	for key, value := range extensions {
		fields = append(fields, reflect.StructField{
			Name: extensionStructFieldName(key),
			Type: reflect.TypeOf(value),
			Tag:  reflect.StructTag(`json:"` + key + `"`),
		})
	}

	structType := reflect.StructOf(fields)
	structValue := reflect.New(structType).Elem()
	structValue.Field(0).Set(reflect.ValueOf(primaryObject))
	for key, value := range extensions {
		structValue.FieldByName(extensionStructFieldName(key)).Set(reflect.ValueOf(value))
	}

	return json.Marshal(structValue.Interface())
}
