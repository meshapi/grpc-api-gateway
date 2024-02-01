package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/meshapi/grpc-rest-gateway/examples/internal/gen/integration"
	"github.com/meshapi/grpc-rest-gateway/gateway"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestQueryParams(t *testing.T) {
	manager := StartSharedTestServer()
	mux := gateway.NewServeMux()
	integration.RegisterQueryParamsTestHandler(context.Background(), mux, manager.ClientConnection())

	tests := []struct {
		Name     string
		Request  *http.Request
		Response string
	}{
		{
			Name: "AutoMapAll-Basic",
			Request: NewRequest("GET", "/query/auto-map-all",
				url.Values{
					"id":         []string{"ID"},
					"num":        []string{"51"},
					"month_name": []string{"Jan"},
				}, nil),
			Response: `{"id":"ID","num":51,"month_name":"Jan"}`,
		},
		{
			Name: "AutoMapAll-EnumValue",
			Request: NewRequest("GET", "/query/auto-map-all",
				url.Values{
					"id":       []string{"ID"},
					"priority": []string{"2"},
				}, nil),
			Response: `{"id":"ID","priority":"Low"}`,
		},
		{
			Name: "AutoMapAll-EnumName",
			Request: NewRequest("GET", "/query/auto-map-all",
				url.Values{
					"id":       []string{"ID"},
					"priority": []string{"Medium"},
				}, nil),
			Response: `{"id":"ID","priority":"Medium"}`,
		},
		{
			Name: "AutoMapAll-OneOf-Nested",
			Request: NewRequest("GET", "/query/auto-map-all",
				url.Values{
					"note_details.type_name": []string{"TypeName"},
					"nested_detail.text":     []string{"NestedDetailText"},
				}, nil),
			Response: `{"note_details":{"type_name":"TypeName"},"nested_detail":{"text":"NestedDetailText"}}`,
		},
		{
			Name: "AutoMapAll-Repeated-And-Table",
			Request: NewRequest("GET", "/query/auto-map-all",
				url.Values{
					"repeated_strings": []string{"a", "b", "c", "d"},
					"table[key1]":      []string{"value1"},
					"table[key2]":      []string{"value2"},
				}, nil),
			Response: `{"repeated_strings":["a","b","c","d"],"table":{"key1":"value1","key2":"value2"}}`,
		},
		{
			Name: "AliasOverride",
			Request: NewRequest("GET", "/query/alias-override",
				url.Values{
					"identification": []string{"ID"},
					"note_text":      []string{"NoteDetailsText"},
					"table[key]":     []string{"value"},
				}, nil),
			Response: `{"id":"ID","note_details":{"text":"NoteDetailsText"},"table":{"key":"value"}}`,
		},
		{
			// when creating an alias, older selector should not get recognized anymore.
			Name:     "AliasOverride-RemoveOriginalSelector",
			Request:  NewRequest("GET", "/query/alias-override", url.Values{"note_details.text": []string{"whatever"}}, nil),
			Response: `{}`,
		},
		{
			Name: "IgnoreFields",
			Request: NewRequest("GET", "/query/ignore-fields",
				url.Values{
					"id":                 []string{"ID"},    // should be no-op
					"note_details.text":  []string{"Text1"}, // should be no-op
					"table[key]":         []string{"value"},
					"nested_detail.text": []string{"Text2"},
				}, nil),
			Response: `{"table":{"key":"value"}, "nested_detail":{"text":"Text2"}}`,
		},
		{
			Name: "DisableAllParams",
			Request: NewRequest("GET", "/query/disable-all-params",
				url.Values{
					"id":                 []string{"ID"},
					"num":                []string{"1"},
					"repeated_strings":   []string{"a"},
					"note_details.text":  []string{"Text1"},
					"table[key]":         []string{"value"},
					"nested_detail.text": []string{"Text2"},
				}, nil),
			Response: `{}`,
		},
		{
			Name: "AliasesOnly",
			Request: NewRequest("GET", "/query/aliases-only",
				url.Values{
					"id":                 []string{"ID"},
					"type_id":            []string{"1"},
					"num":                []string{"1"},
					"repeated_strings":   []string{"a"},
					"note_details.text":  []string{"Text1"},
					"table[key]":         []string{"value"},
					"nested_detail.text": []string{"Text2"},
				}, nil),
			Response: `{"id":"ID","nested_detail":{"type_id":1}}`,
		},
		{
			Name: "BodyAndParams",
			Request: NewRequest("POST", "/query/body-and-params/12",
				url.Values{
					"id":               []string{"ID"}, // should not get used.
					"repeated_strings": []string{"a", "b"},
				}, strings.NewReader(`{"text":"Text"}`)),
			Response: `{"id":"12","repeated_strings":["a","b"],"nested_detail":{"text":"Text"}}`,
		},
		{
			Name: "BodyOnly",
			Request: NewRequest("POST", "/query/body-only",
				// no params should get used
				url.Values{
					"id":               []string{"ID"},
					"repeated_strings": []string{"a", "b"},
				}, strings.NewReader(`{"id":"1"}`)),
			Response: `{"id":"1"}`,
		},
		{
			Name: "CustomMethod",
			Request: NewRequest("TEST", "/query/auto-map-all",
				url.Values{
					"id":               []string{"ID"},
					"repeated_strings": []string{"a", "b"},
				}, nil),
			Response: `{"id":"ID","repeated_strings":["a","b"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			mux.ServeHTTP(responseRecorder, tt.Request)

			if responseRecorder.Result().StatusCode != 200 {
				t.Fatalf("received status code %d", responseRecorder.Result().StatusCode)
				return
			}

			expectedResponse := &integration.TestMessage{}
			if !Unmarshal(t, strings.NewReader(tt.Response), expectedResponse) {
				return
			}

			body := responseRecorder.Result().Body
			defer body.Close()

			response := &integration.TestMessage{}
			if !Unmarshal(t, body, response) {
				return
			}

			if diff := cmp.Diff(expectedResponse, response, protocmp.Transform()); diff != "" {
				t.Fatalf("incorrect response:\n%s", diff)
				return
			}
		})
	}
}
