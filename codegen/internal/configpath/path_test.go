package configpath_test

import (
	"testing"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/configpath"
)

func TestBuild(t *testing.T) {
	testCases := []struct {
		Name    string
		Pattern string
		Input   string
		Expect  string
	}{
		{
			Name:    "Dir",
			Pattern: "{{ .Dir }}/gateway",
			Input:   "path/to/file.proto",
			Expect:  "path/to/gateway",
		},
		{
			Name:    "Name",
			Pattern: "{{ .Name }}_gateway",
			Input:   "path/to/file.proto",
			Expect:  "file_gateway",
		},
		{
			Name:    "Path",
			Pattern: "{{ .Path }}_gateway",
			Input:   "path/to/file.proto",
			Expect:  "path/to/file_gateway",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			builder, err := configpath.NewBuilder(tt.Pattern)
			if err != nil {
				t.Fatalf("received unexpected error building the pattern: %s", err)
				return
			}

			result, err := builder.Build(tt.Input)
			if err != nil {
				t.Fatalf("received unexpected error building result: %s", err)
				return
			}

			if result != tt.Expect {
				t.Fatalf("expected %v but received %v", tt.Expect, result)
			}
		})
	}
}

func Benchmark(b *testing.B) {
	x, _ := configpath.NewBuilder("{{ .Path }}_gateway")
	for i := 0; i < b.N; i++ {
		result, _ := x.Build("path/to/file.proto")
		if result != "path/to/file_gateway" {
			b.Fail()
		}
	}
}
