package httprule_test

import (
	"strings"
	"testing"

	"github.com/meshapi/grpc-api-gateway/pkg/httprule"
)

func TestParse(t *testing.T) {
	tests := []struct {
		Path              string
		ExpectedOutput    string
		ExpectedPattern   string
		ExpectedErrString string
	}{
		{
			Path:              "",
			ExpectedErrString: "invalid HTTP rule, no leading /",
		},
		{
			Path:            "/",
			ExpectedOutput:  "/",
			ExpectedPattern: "/",
		},
		{
			Path:            "/v1/echo/",
			ExpectedOutput:  "/v1/echo",
			ExpectedPattern: "/v1/echo",
		},
		{
			Path:            "/v1/echo///",
			ExpectedOutput:  "/v1/echo",
			ExpectedPattern: "/v1/echo",
		},
		{
			Path:            "/v1/echo/{data}",
			ExpectedOutput:  "/v1/echo/['data']",
			ExpectedPattern: "/v1/echo/?",
		},
		{
			Path:            "/v1/echo/{data}/{request.label}",
			ExpectedOutput:  "/v1/echo/['data']/['request.label']",
			ExpectedPattern: "/v1/echo/?/?",
		},
		{
			Path:            "/v1/echo/{data}/{request.label}/{rest=*}",
			ExpectedOutput:  "/v1/echo/['data']/['request.label']/[*'rest']",
			ExpectedPattern: "/v1/echo/?/?/*",
		},
		{
			Path:            "/v*/echo",
			ExpectedOutput:  "/v*/echo",
			ExpectedPattern: "/v*/echo",
		},
		{
			Path:              "/v|",
			ExpectedErrString: "invalid literal segment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Path, func(t *testing.T) {
			tpl, err := httprule.Parse(tt.Path)
			if err != nil {
				if tt.ExpectedErrString == "" {
					t.Fatalf("unexpected error: %s", err)
				} else if !strings.Contains(err.Error(), tt.ExpectedErrString) {
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

			output = tpl.Pattern()
			if output != tt.ExpectedPattern {
				t.Fatalf("expected '%s', received '%s'", tt.ExpectedPattern, output)
			}
		})
	}
}
