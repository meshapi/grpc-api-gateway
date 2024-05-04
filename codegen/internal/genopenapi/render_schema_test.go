package genopenapi

import (
	"testing"

	"google.golang.org/genproto/googleapis/api/visibility"
)

func TestVisibilityCheck(t *testing.T) {
	externalMap := Generator{
		Options: Options{
			VisibilitySelectors: SelectorMap{
				"external": true,
			},
		},
	}

	testCases := []struct {
		Input  string
		Result bool
	}{
		{Input: "internal,api,support", Result: false},
		{Input: "internal,api,support,external", Result: true},
		{Input: "", Result: true},
	}

	for _, tt := range testCases {
		t.Run(tt.Input, func(t *testing.T) {
			if tt.Result != externalMap.isVisible(&visibility.VisibilityRule{Restriction: tt.Input}) {
				t.Fatalf("expected %v but received %v", tt.Result, !tt.Result)
			}
		})
	}
}
