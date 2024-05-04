package gateway

import (
	"net/url"
	"strings"

	"github.com/meshapi/grpc-api-gateway/dotpath"
	"github.com/meshapi/grpc-api-gateway/protopath"
	"github.com/meshapi/grpc-api-gateway/trie"
	"google.golang.org/protobuf/proto"
)

// QueryParameterParseOptions hold all inputs for parsing query parameters.
type QueryParameterParseOptions struct {
	// Filter holds a trie that can block already bound or otherwise ignored query paramters.
	Filter *trie.Node

	// Aliases is a table of arbitrary names mapped to field keys.
	Aliases map[string]string

	// LimitToAliases limits the query parameters to aliases only. Used when auto discovery is disabled.
	LimitToAliases bool
}

// QueryParameterParser defines interface for all query parameter parsers.
type QueryParameterParser interface {
	Parse(msg proto.Message, values url.Values, inputs QueryParameterParseOptions) error
}

// PopulateQueryParameters parses query parameters
// into "msg" using current query parser.
func (s *ServeMux) PopulateQueryParameters(msg proto.Message, values url.Values, inputs QueryParameterParseOptions) error {
	return s.queryParamParser.Parse(msg, values, inputs)
}

// DefaultQueryParser is a QueryParameterParser which implements the default
// query parameters parsing behavior.
//
// See https://github.com/grpc-ecosystem/grpc-gateway/issues/2632 for more context.
type DefaultQueryParser struct{}

// Parse populates "values" into "msg".
// A value is ignored if its key starts with one of the elements in "filter".
func (*DefaultQueryParser) Parse(msg proto.Message, values url.Values, input QueryParameterParseOptions) error {
	for key, values := range values {
		if messageKey, mapKey, ok := matchMapKey(key); ok {
			key = messageKey
			values = append([]string{mapKey}, values...)
		}
		fieldKey, ok := input.Aliases[key]
		if ok {
			key = fieldKey
		} else if input.LimitToAliases {
			continue
		}
		fieldPath := dotpath.Parse(&key)
		if !ok && input.Filter.HasCommonPrefix(fieldPath) {
			continue
		}
		if err := protopath.PopulateFieldValueFromPath(msg.ProtoReflect(), fieldPath, values); err != nil {
			return err
		}
	}
	return nil
}

func matchMapKey(key string) (string, string, bool) {
	start := strings.IndexByte(key, '[')
	if start == -1 { // 0 is also not acceptable because it means there is no key, only braces.
		return "", "", false
	}
	end := strings.LastIndexByte(key, ']')
	if end == -1 || end != len(key)-1 || end <= start {
		return "", "", false
	}
	return key[:start], key[start+1 : end], true
}
