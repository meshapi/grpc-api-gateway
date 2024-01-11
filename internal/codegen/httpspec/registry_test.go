package httpspec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/meshapi/grpc-rest-gateway/internal/codegen/httpspec"
)

const (
	testConfigYAML = `
gateway:
 endpoints:
  - selector: 'meshapi.example.v1.Test'
    get: '/v1/test'
    additional_bindings:
    - post: '/v1/test'
      body: 'data'
      disable_query_param_discovery: true
    - custom:
        path: '/v1/test'
        method: 'REGISTER'

  - selector: 'meshapi.example.v2.Test'
    post: '/v1/test'
    body: 'data'
    query_params:
      - selector: 'filters.show_all'
        name: 'show_all'
`

	testConfigJSON = `
{
    "gateway": {
        "endpoints": [
            {
                "selector": "meshapi.example.v1.Test",
                "get": "/v1/test",
                "additional_bindings": [
                    {
                        "post": "/v1/test",
                        "body": "data",
                        "disable_query_param_discovery": true
                    },
                    {
                        "custom": {
                            "path": "/v1/test",
                            "method": "REGISTER"
                        }
                    }
                ]
            },
            {
                "selector": "meshapi.example.v2.Test",
                "post": "/v1/test",
                "body": "data",
                "query_params": [
                    {
                        "selector": "filters.show_all",
                        "name": "show_all"
                    }
                ]
            }
        ]
    }
}`
)

func TestLoadYAML(t *testing.T) {
	r := httpspec.NewRegistry()
	testDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(testDir, "config.yaml"), []byte(testConfigYAML), 07555); err != nil {
		t.Fatalf("unexpected failure in writing test data: %s", err)
		return
	}

	if err := os.WriteFile(filepath.Join(testDir, "config.json"), []byte(testConfigJSON), 07555); err != nil {
		t.Fatalf("unexpected failure in writing test data: %s", err)
		return
	}

	if err := r.LoadFromFile(filepath.Join(testDir, "config.yaml"), ""); err != nil {
		t.Fatalf("unexpected failure in loading test data: %s", err)
		return
	}

	_, ok := r.LookupBinding(".meshapi.example.v1.Test")
	if !ok {
		t.Fatal("unexpected failure in looking up meshapi.example.v1.Test")
		return
	}
}
