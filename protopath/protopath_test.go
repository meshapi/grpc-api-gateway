package protopath_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/meshapi/grpc-api-gateway/internal/examplepb"
	"github.com/meshapi/grpc-api-gateway/protopath"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestPopulateFromPath(t *testing.T) {
	testCases := []struct {
		Name    string
		Message proto.Message
		Path    string
		Value   string
		Result  proto.Message
		Error   bool
	}{
		{
			Name:    "Enum",
			Message: &examplepb.Proto3Message{},
			Path:    "enum_value",
			Value:   "Y",
			Result:  &examplepb.Proto3Message{EnumValue: examplepb.EnumValue_Y},
		},
		{
			Name:    "Nested",
			Message: &examplepb.Proto3Message{},
			Path:    "nested.double_value",
			Value:   "5.6",
			Result:  &examplepb.Proto3Message{Nested: &examplepb.Proto3Message{DoubleValue: 5.6}},
		},
		{
			Name:    "NestedDeep",
			Message: &examplepb.Proto3Message{},
			Path:    "nested.nested.repeated_enum",
			Value:   "Z",
			Result: &examplepb.Proto3Message{
				Nested: &examplepb.Proto3Message{
					Nested: &examplepb.Proto3Message{
						RepeatedEnum: []examplepb.EnumValue{examplepb.EnumValue_Z},
					},
				},
			},
		},
		{
			Name:    "OneOf",
			Message: &examplepb.Proto3Message{},
			Path:    "oneof_bool_value",
			Value:   "true",
			Result: &examplepb.Proto3Message{
				OneofValue: &examplepb.Proto3Message_OneofBoolValue{OneofBoolValue: true},
			},
		},
		{
			Name: "OneOf",
			Message: &examplepb.Proto3Message{
				OneofValue: &examplepb.Proto3Message_OneofBoolValue{OneofBoolValue: true},
			},
			Path:  "oneof_bool_value",
			Value: "true",
			Result: &examplepb.Proto3Message{
				OneofValue: &examplepb.Proto3Message_OneofStringValue{OneofStringValue: "string_value"},
			},
			Error: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := protopath.PopulateFieldFromPath(tt.Message, tt.Path, tt.Value); err != nil {
				if !tt.Error {
					t.Fatalf("received error when expected no error: %v", err)
				}
				return
			}
			if tt.Error {
				t.Fatal("expected error but received none")
				return
			}

			if diff := cmp.Diff(tt.Message, tt.Result, protocmp.Transform()); diff != "" {
				t.Fatalf("unexpected result: %s", diff)
			}
		})
	}
}
