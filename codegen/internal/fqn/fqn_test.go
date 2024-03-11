package fqn_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/fqn"
)

func TestFQN(t *testing.T) {
	testCases := []struct {
		Input string
	}{
		{Input: ""},
		{Input: "."},
		{Input: ".."},
		{Input: "a.b.c"},
		{Input: ".a.b.c"},
		{Input: ".a.b.c."},
		{Input: "a.b.c."},
	}

	for _, tt := range testCases {
		t.Run(tt.Input, func(t *testing.T) {
			item := fqn.Parse(&tt.Input)

			stringParts := strings.Split(tt.Input, ".")

			fqnParts := item.Parts()
			if !reflect.DeepEqual(fqnParts, stringParts) {
				t.Fatalf("expected %+v, received %+v", stringParts, fqnParts)
			}

			expectedMaxDepth := strings.Count(tt.Input, ".")
			if item.MaxDepth() != expectedMaxDepth {
				t.Fatalf("expected max depth %d, received: %d", expectedMaxDepth, item.MaxDepth())
			}

			// for every string, say a.b.c, where parts are [a b c]
			// at each index, joining all items from that index to the end should
			// create expected depth L-1-index. For instance, at index 1, b.c would be
			// expected string at depth of 3-1-1 = 1.
			for i := range stringParts {
				depth := len(stringParts) - 1 - i

				fqnParts := item.PartsAtDepth(depth)
				if !reflect.DeepEqual(fqnParts, stringParts[i:]) {
					t.Fatalf("expected[depth=%d] %+v, received %+v", depth, stringParts, fqnParts)
				}

				expectedStringAtDepth := strings.Join(stringParts[i:], ".")
				receivedString := item.StringAtDepth(depth)
				if expectedStringAtDepth != receivedString {
					t.Fatalf("expected[depth=%d] %s, received %s", depth, expectedStringAtDepth, receivedString)
				}
			}
		})
	}
}
