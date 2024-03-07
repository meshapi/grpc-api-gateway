package openapiv3_test

import (
	"errors"
	"testing"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
)

type Root struct {
	A string `validate:"required"`
	B string
	C []Nested `validate:"required"`
	D []Nested
	E map[string]Nested `validate:"required"`
	F map[string]Nested
	G Custom `validate:"required"`
}

type Nested struct {
	Name string `validate:"required"`
}

type Custom struct {
	Fail bool
}

func (c *Custom) Validate(*openapiv3.Document) error {
	if c.Fail {
		return errors.New("failed")
	}

	return nil
}

func TestValidation(t *testing.T) {

	testCases := []struct {
		Name  string
		Value Root
		Error string
	}{
		{
			Name: "A-required-err",
			Value: Root{
				A: "",
				C: []Nested{{Name: "Some"}},
				E: map[string]Nested{
					"A": {Name: "Some"},
				},
				G: Custom{Fail: false}},
			Error: `"A" is required`,
		},
		{
			Name: "A-required-ok",
			Value: Root{
				A: "Satisfied",
				C: []Nested{{Name: "Some"}},
				E: map[string]Nested{
					"A": {Name: "Some"},
				},
				G: Custom{Fail: false}},
		},
		{
			Name: "C-required-err",
			Value: Root{
				A: "Some",
				C: []Nested{},
				E: map[string]Nested{
					"A": {Name: "Some"},
				},
				G: Custom{Fail: false}},
			Error: `"C" is required`,
		},
		{
			Name: "C-required-ok",
			Value: Root{
				A: "Some",
				C: []Nested{{Name: "Some"}},
				E: map[string]Nested{
					"A": {Name: "Some"},
				},
				G: Custom{Fail: false}},
		},
		{
			Name: "E-required-err",
			Value: Root{
				A: "Some",
				C: []Nested{{Name: "Some"}},
				E: map[string]Nested{},
				G: Custom{Fail: false}},
			Error: `"E" is required`,
		},
		{
			Name: "E-required-ok",
			Value: Root{
				A: "Some",
				C: []Nested{{Name: "Some"}},
				E: map[string]Nested{
					"A": {Name: "Some"},
				},
				G: Custom{Fail: false}},
		},
		{
			Name: "G-custom-validate-err",
			Value: Root{
				A: "Some",
				C: []Nested{{Name: "Some"}},
				E: map[string]Nested{
					"a": {Name: "Some"},
				},
				G: Custom{Fail: true}},
			Error: `invalid "G": failed`,
		},
		{
			Name: "G-custom-validate-ok",
			Value: Root{
				A: "Some",
				C: []Nested{{Name: "Some"}},
				E: map[string]Nested{
					"A": {Name: "Some"},
				},
				G: Custom{Fail: false}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			err := openapiv3.Validate(&tt.Value, nil)
			if err == nil {
				if tt.Error != "" {
					t.Fatalf("expected error %q but received none", tt.Error)
					return
				}
			} else if err.Error() != tt.Error {
				t.Fatalf("expected error %q but received %q", tt.Error, err)
			}
		})
	}
}
