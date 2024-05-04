package dotpath_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/meshapi/grpc-api-gateway/dotpath"
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
			item := dotpath.Parse(&tt.Input)

			stringParts := strings.Split(tt.Input, ".")

			fqnParts := item.Parts()
			if !reflect.DeepEqual(fqnParts, stringParts) {
				t.Fatalf("expected %+v, received %+v", stringParts, fqnParts)
			}

			for index, part := range stringParts {
				if result := item.Index(index); result != part {
					t.Fatalf("expected[index=%d] %+v, received %+v", index, part, result)
				}
			}

			expectedMaxDepth := strings.Count(tt.Input, ".")
			if item.MaxDepth() != expectedMaxDepth {
				t.Fatalf("expected max depth %d, received: %d", expectedMaxDepth, item.MaxDepth())
			}

			// for every string, say a.b.c, where parts are [a b c]
			// at each index, joining all items from that index to the end should
			// create expected depth L-1-index. For instance, at index 1, b.c would be
			// the expected string at depth of 3-1-1 = 1.
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

				expectedTrimmedString := strings.Join(stringParts[:len(stringParts)-i], ".")
				trimmedString := item.TrimmedSuffix(i)
				if trimmedString != expectedTrimmedString {
					t.Fatalf(
						"trimmed string at [n=%d] should be %s but is %s", i, trimmedString, receivedString)
				}
			}
		})
	}
}

func BenchmarkSplitAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = strings.Split("a.b.c.d.e.f", ".")[1]
	}
}

func BenchmarkParts(b *testing.B) {
	str := "a.b.c.d.e.f"
	for i := 0; i < b.N; i++ {
		dotpath.Parse(&str).Index(1)
	}
}
