package gengateway

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/meshapi/grpc-rest-gateway/internal/casing"
	"github.com/meshapi/grpc-rest-gateway/internal/codegen/descriptor"
	"github.com/meshapi/grpc-rest-gateway/internal/httprule"
	"github.com/meshapi/grpc-rest-gateway/trie"
)

type param struct {
	*descriptor.File
	Imports            []descriptor.GoPackage
	UseRequestContext  bool
	RegisterFuncSuffix string
	AllowPatchFeature  bool
	OmitPackageDoc     bool
}

type binding struct {
	*descriptor.Binding
	Registry          *descriptor.Registry
	AllowPatchFeature bool
}

// GetBodyFieldPath returns the binding body's field path.
func (b binding) GetBodyFieldPath() string {
	if b.Body != nil && len(b.Body.FieldPath) != 0 {
		return b.Body.FieldPath.String()
	}
	return "*"
}

// GetBodyFieldStructName returns the binding body's struct field name.
func (b binding) GetBodyFieldStructName() (string, error) {
	if b.Body != nil && len(b.Body.FieldPath) != 0 {
		return casing.Camel(b.Body.FieldPath.String()), nil
	}
	return "", errors.New("no body field found")
}

func (b binding) QueryParameterFilter() queryParameterFilter {
	if b.QueryParameterCustomization.DisableAutoDiscovery {
		return queryParameterFilter{DoubleArray: trie.New(nil)}
	}
	return queryParameterFilter{DoubleArray: b.Binding.QueryParameterFilter()}
}

// HasEnumPathParam returns true if the path parameter slice contains a parameter
// that maps to an enum proto field that is not repeated, if not false is returned.
func (b binding) HasEnumPathParam() bool {
	return b.hasEnumPathParam(false)
}

// HasRepeatedEnumPathParam returns true if the path parameter slice contains a parameter
// that maps to a repeated enum proto field, if not false is returned.
func (b binding) HasRepeatedEnumPathParam() bool {
	return b.hasEnumPathParam(true)
}

// hasEnumPathParam returns true if the path parameter slice contains a parameter
// that maps to a enum proto field and that the enum proto field is or isn't repeated
// based on the provided 'repeated' parameter.
func (b binding) hasEnumPathParam(repeated bool) bool {
	for _, p := range b.PathParameters {
		if p.IsEnum() && p.IsRepeated() == repeated {
			return true
		}
	}
	return false
}

// LookupEnum looks up a enum type by path parameter.
func (b binding) LookupEnum(p descriptor.Parameter) *descriptor.Enum {
	e, err := b.Registry.LookupEnum("", p.Target.GetTypeName())
	if err != nil {
		return nil
	}
	return e
}

// FieldMaskField returns the golang-style name of the variable for a FieldMask, if there is exactly one of that type in
// the message. Otherwise, it returns an empty string.
func (b binding) FieldMaskField() string {
	var fieldMaskField *descriptor.Field
	for _, f := range b.Method.RequestType.Fields {
		if f.GetTypeName() == ".google.protobuf.FieldMask" {
			// if there is more than 1 FieldMask for this request, then return none
			if fieldMaskField != nil {
				return ""
			}
			fieldMaskField = f
		}
	}
	if fieldMaskField != nil {
		return casing.Camel(fieldMaskField.GetName())
	}
	return ""
}

// queryParameterFilter is a wrapper of trie.DoubleArray which provides String() to output DoubleArray.Encoding in a stable and predictable format.
type queryParameterFilter struct {
	*trie.DoubleArray
}

func (f queryParameterFilter) String() string {
	encodings := make([]string, len(f.Encoding))
	for str, enc := range f.Encoding {
		encodings[enc] = fmt.Sprintf("%q: %d", str, enc)
	}
	e := strings.Join(encodings, ", ")
	return fmt.Sprintf("&trie.DoubleArray{Encoding: map[string]int{%s}, Base: %#v, Check: %#v}", e, f.Base, f.Check)
}

func prepareHTTPPath(path *httprule.Template) string {
	writer := &strings.Builder{}

	if len(path.Segments) == 0 {
		return "/"
	}

	for index, segment := range path.Segments {
		switch segment.Type {
		case httprule.SegmentTypeLiteral:
			_, _ = fmt.Fprintf(writer, "/%s", segment.Value)
		case httprule.SegmentTypeSelector:
			_, _ = fmt.Fprintf(writer, "/:%s", segment.Value)
		case httprule.SegmentTypeWildcard:
			_, _ = fmt.Fprintf(writer, "/:_segment_"+strconv.Itoa(index))
		case httprule.SegmentTypeCatchAllSelector:
			_, _ = fmt.Fprintf(writer, "/*%s", segment.Value)
		default:
			_, _ = fmt.Fprintf(writer, "/<!?:%s>", segment.Value)
		}
	}

	return writer.String()
}

func prepareHTTPPattern(path *httprule.Template) string {
	writer := &strings.Builder{}

	if len(path.Segments) == 0 {
		return "/"
	}

	for _, segment := range path.Segments {
		switch segment.Type {
		case httprule.SegmentTypeLiteral:
			_, _ = fmt.Fprintf(writer, "/%s", segment.Value)
		case httprule.SegmentTypeSelector:
			_, _ = fmt.Fprintf(writer, "/{%s}", segment.Value)
		case httprule.SegmentTypeWildcard:
			_, _ = fmt.Fprintf(writer, "/?")
		case httprule.SegmentTypeCatchAllSelector:
			_, _ = fmt.Fprintf(writer, "/{%s=*}", segment.Value)
		default:
			_, _ = fmt.Fprintf(writer, "/<!?:%s>", segment.Value)
		}
	}

	return writer.String()

}

type trailerParams struct {
	Services           []*descriptor.Service
	UseRequestContext  bool
	RegisterFuncSuffix string
}

func (g *Generator) applyTemplate(p param, reg *descriptor.Registry) (string, error) {
	w := bytes.NewBuffer(nil)
	if err := headerTemplate.Execute(w, p); err != nil {
		return "", err
	}
	var targetServices []*descriptor.Service

	for _, msg := range p.Messages {
		msgName := casing.Camel(*msg.Name)
		msg.Name = &msgName
	}

	for _, svc := range p.Services {
		var methodWithBindingsSeen bool
		svcName := casing.Camel(*svc.Name)
		svc.Name = &svcName

		for _, meth := range svc.Methods {
			methName := casing.Camel(*meth.Name)
			meth.Name = &methName
			for _, b := range meth.Bindings {
				if err := g.CheckDuplicateEndpoint(b.HTTPMethod, b.PathTemplate.Pattern(), svc); err != nil {
					return "", err
				}

				methodWithBindingsSeen = true
				if err := handlerTemplate.Execute(w, binding{
					Binding:           b,
					Registry:          reg,
					AllowPatchFeature: p.AllowPatchFeature,
				}); err != nil {
					return "", err
				}

				// Local
				if false {
					if err := localHandlerTemplate.Execute(w, binding{
						Binding:           b,
						Registry:          reg,
						AllowPatchFeature: p.AllowPatchFeature,
					}); err != nil {
						return "", err
					}
				}
			}
		}
		if methodWithBindingsSeen {
			targetServices = append(targetServices, svc)
		}
	}
	if len(targetServices) == 0 {
		return "", nil
	}

	tp := trailerParams{
		Services:           targetServices,
		UseRequestContext:  p.UseRequestContext,
		RegisterFuncSuffix: p.RegisterFuncSuffix,
	}
	// Local
	if false {
		if err := localTrailerTemplate.Execute(w, tp); err != nil {
			return "", err
		}
	}

	if err := trailerTemplate.Execute(w, tp); err != nil {
		return "", err
	}
	return w.String(), nil
}

var (
	//go:embed templates/header.tmpl
	templateDataHeader string
	headerTemplate     = template.Must(template.New("header").Parse(templateDataHeader))

	//go:embed templates/handler.tmpl
	templateDataHandler string
	handlerTemplate     = template.Must(template.New("handler").Parse(templateDataHandler))

	//go:embed templates/request_func_signature.tmpl
	templateDataRequestFuncSignature string
	_                                = template.Must(
		handlerTemplate.New("request-func-signature").Parse(
			strings.ReplaceAll(templateDataRequestFuncSignature, "\n", "")))

	//go:embed templates/client_streaming_request_func.tmpl
	templateDataClientStreamingRequestFunc string
	_                                      = template.Must(
		handlerTemplate.New("client-streaming-request-func").Parse(templateDataClientStreamingRequestFunc))

	funcMap template.FuncMap = map[string]interface{}{
		"camelIdentifier": casing.CamelIdentifier,
	}

	//go:embed templates/client_rpc_request_func.tmpl
	templateDataClientRPCRequestFunc string
	_                                = template.Must(
		handlerTemplate.New("client-rpc-request-func").Funcs(funcMap).Parse(templateDataClientRPCRequestFunc))

	//go:embed templates/bidi_streaming_request_func.tmpl
	templateDataBiDiStreamingRequestFunc string
	_                                    = template.Must(
		handlerTemplate.New("bidi-streaming-request-func").Parse(templateDataBiDiStreamingRequestFunc))

	//go:embed templates/local_handler.tmpl
	templateDataLocalHandler string
	localHandlerTemplate     = template.Must(template.New("local-handler").Parse(templateDataLocalHandler))

	//go:embed templates/local_request_func_signature.tmpl
	templateDataLocalRequestFuncSignature string
	_                                     = template.Must(
		localHandlerTemplate.New("local-request-func-signature").Parse(
			strings.ReplaceAll(templateDataLocalRequestFuncSignature, "\n", "")))

	//go:embed templates/local_client_request_func.tmpl
	templateDataLocalClientRPCRequestFunc string
	_                                     = template.Must(
		localHandlerTemplate.New("local-client-rpc-request-func").Funcs(funcMap).Parse(
			templateDataLocalClientRPCRequestFunc))

	//go:embed templates/local_trailer.tmpl
	templateDataLocalTrailer string
	localTrailerTemplate     = template.Must(template.New("local-trailer").Parse(templateDataLocalTrailer))

	//go:embed templates/trailer.tmpl
	templateDataTrailer string
	trailerFuncMap      = map[string]interface{}{
		"httpPath":    prepareHTTPPath,
		"httpPattern": prepareHTTPPattern,
	}
	trailerTemplate = template.Must(template.New("trailer").Funcs(trailerFuncMap).Parse(templateDataTrailer))
)
