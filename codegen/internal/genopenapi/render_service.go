package genopenapi

import (
	"fmt"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"github.com/meshapi/grpc-rest-gateway/pkg/httprule"
)

func (g *Generator) renderOperation(binding *descriptor.Binding) (*openapiv3.OperationCore, error) {
	operation := &openapiv3.OperationCore{}

	if !g.DisableServiceTags {
		tag := binding.Method.Service.GetName()
		if g.IncludePackageInTags {
			if pkg := binding.Method.Service.File.GetPackage(); pkg != "" {
				tag = pkg + "." + tag
			}
		}

		operation.Tags = append(operation.Tags, tag)
	}

	return operation, nil
}

func (r *Registry) renderPath(binding *descriptor.Binding) string {
	writer := &strings.Builder{}

	if len(binding.PathTemplate.Segments) == 0 {
		return "/"
	}

	for _, segment := range binding.PathTemplate.Segments {
		switch segment.Type {
		case httprule.SegmentTypeLiteral:
			writer.WriteString("/" + segment.Value)
		case httprule.SegmentTypeSelector:
			writer.WriteString("/{" + segment.Value + "}")
		case httprule.SegmentTypeWildcard:
			_, _ = fmt.Fprintf(writer, "/?")
		case httprule.SegmentTypeCatchAllSelector:
			writer.WriteString("/{" + segment.Value + "}")
		default:
			_, _ = fmt.Fprintf(writer, "/<!?:%s>", segment.Value)
		}
	}

	return writer.String()
}
