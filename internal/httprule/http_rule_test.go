package httprule_test

import (
	"strings"
	"testing"

	"github.com/meshapi/grpc-rest-gateway/internal/httprule"
)

func TestParse(t *testing.T) {
	tests := []struct {
		Path              string
		ExpectedOutput    string
		ExpectedErrString string
	}{
		{
			Path:              "",
			ExpectedErrString: "invalid HTTP rule, missing leading /",
		},
		{
			Path:           "/",
			ExpectedOutput: "/",
		},
		{
			Path:           "/v1/echo/",
			ExpectedOutput: "/v1/echo",
		},
		{
			Path:           "/v1/echo///",
			ExpectedOutput: "/v1/echo",
		},
		{
			Path:           "/v1/echo/{data}",
			ExpectedOutput: "/v1/echo/['data']",
		},
		{
			Path:           "/v1/echo/{data}/{request.label}",
			ExpectedOutput: "/v1/echo/['data']/['request.label']",
		},
		{
			Path:           "/v1/echo/{data}/{request.label}/{rest=*}",
			ExpectedOutput: "/v1/echo/['data']/['request.label']/[*'rest']",
		},
		{
			Path:           "/v*/echo",
			ExpectedOutput: "/v*/echo",
		},
		{
			Path:           "/v1/*/something",
			ExpectedOutput: "/v1/*/something",
		},
		{
			Path:              "/v|",
			ExpectedErrString: "invalid literal segment",
		},
	}

	for _, tt := range tests {
		tpl, err := httprule.Parse(tt.Path)
		if err != nil {
			if tt.ExpectedErrString == "" {
				t.Fatalf("unexpected error: %s", err)
			} else if strings.Contains(err.Error(), tt.ExpectedErrString) {
				t.Fatalf("expected error containing '%s' but received: %s", tt.ExpectedErrString, err)
			}
			return
		} else if tt.ExpectedErrString != "" {
			t.Fatalf("expected error containing '%s' but received nil", tt.ExpectedErrString)
			return
		}

		output := tpl.String()
		if output != tt.ExpectedOutput {
			t.Fatalf("expected '%s', received '%s'", tt.ExpectedOutput, output)
		}
	}
}
