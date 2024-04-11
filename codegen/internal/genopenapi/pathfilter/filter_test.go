package pathfilter_test

import (
	"testing"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/pathfilter"
)

func TestFilter(t *testing.T) {
	items := [...]string{
		"a.b.c",
		"a.b.d.e",
		"x.y.z",
	}
	filter := pathfilter.New()
	for _, item := range items {
		filter.PutString(item)
	}

	testCases := []struct {
		Prefix      string
		Has         bool
		ChildPrefix string
		ChildHas    bool
	}{
		{
			Prefix:      "a",
			Has:         true,
			ChildPrefix: "b.c",
			ChildHas:    true,
		},
		{
			Prefix:      "a.b.c",
			Has:         true,
			ChildPrefix: "x",
			ChildHas:    false,
		},
		{
			Prefix:      "a",
			Has:         true,
			ChildPrefix: "d",
			ChildHas:    false,
		},
		{
			Prefix: "a.d",
			Has:    false,
		},
		{
			Prefix: "x.y.z",
			Has:    true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Prefix, func(t *testing.T) {
			has, child := filter.HasString(tt.Prefix)
			if tt.Has != has {
				t.Fatalf("expected %v, received %v", tt.Has, has)
			}
			if !has {
				return
			}

			if tt.ChildHas && child == nil {
				t.Fatal("child instance is not available")
			}
			if has, _ := child.HasString(tt.ChildPrefix); has != tt.ChildHas {
				t.Fatalf("expected child %v, received %v", tt.ChildHas, !tt.ChildHas)
			}
		})
	}
}
