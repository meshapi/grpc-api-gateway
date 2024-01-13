package httprule

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	selectorPattern = regexp.MustCompile(`^\{(\w+(?:[.]\w+)*)\}$`)
	catchAllPattern = regexp.MustCompile(`^\{(\w+(?:[.]\w+)*)=\*\}$`)
	literalPattern  = regexp.MustCompile(`^[a-zA-Z0-9-_.~!@$&'()*+,;=:]+$`)
)

// SegmentType is the type of path segment in the HTTP rule.
type SegmentType uint8

const (
	SegmentTypeLiteral          SegmentType = iota // SegmentTypeLiteral is a literal string.
	SegmentTypeSelector                            // SegmentTypeSelector is a variable.
	SegmentTypeWildcard                            // SegmentTypeWildcard matches a single route segment.
	SegmentTypeCatchAllSelector                    // SegmentTypeCatchAllSelector is similar to selector but reads everything.
)

// Segment is an HTTP rule path segment.
type Segment struct {
	// Value holds string value and is contextualized based on segment type.
	Value string

	// Type describes the type of HTTP rule segment.
	Type SegmentType
}

// Template is an HTTP rule template that can be used to build out routes.
type Template struct {
	// Segments contains all of the path segments.
	Segments []Segment
}

func (t Template) String() string {
	writer := &strings.Builder{}

	if len(t.Segments) == 0 {
		return "/"
	}

	for _, segment := range t.Segments {
		switch segment.Type {
		case SegmentTypeLiteral:
			_, _ = fmt.Fprintf(writer, "/%s", segment.Value)
		case SegmentTypeSelector:
			_, _ = fmt.Fprintf(writer, "/['%s']", segment.Value)
		case SegmentTypeWildcard:
			_, _ = fmt.Fprintf(writer, "/*")
		case SegmentTypeCatchAllSelector:
			_, _ = fmt.Fprintf(writer, "/[*'%s']", segment.Value)
		default:
			_, _ = fmt.Fprintf(writer, "/<!?:%s>", segment.Value)
		}
	}

	return writer.String()
}

// Pattern returns a string representation without variable names. Can be used to compare two templates.
func (t Template) Pattern() string {
	writer := &strings.Builder{}

	if len(t.Segments) == 0 {
		return "/"
	}

	for _, segment := range t.Segments {
		switch segment.Type {
		case SegmentTypeLiteral:
			_, _ = fmt.Fprintf(writer, "/%s", segment.Value)
		case SegmentTypeSelector, SegmentTypeWildcard:
			_, _ = fmt.Fprint(writer, "/?")
		case SegmentTypeCatchAllSelector:
			_, _ = fmt.Fprint(writer, "/*")
		default:
			_, _ = fmt.Fprint(writer, "/<!?>")
		}
	}

	return writer.String()

}

// HasVariables returns whether or not the current template contains any binding variables.
func (t Template) HasVariables() bool {
	for _, segment := range t.Segments {
		if segment.Type == SegmentTypeCatchAllSelector || segment.Type == SegmentTypeSelector {
			return true
		}
	}
	return false
}

// Parse parses an HTTP rule syntax and builds out a template.
func Parse(path string) (Template, error) {
	result := Template{}

	if !strings.HasPrefix(path, "/") {
		return result, fmt.Errorf("invalid HTTP rule, no leading /: %s", path)
	}

	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if segment == "" {
			continue
		}

		switch {
		case strings.HasSuffix(segment, "=*}"):
			matches := catchAllPattern.FindStringSubmatch(segment)
			if matches == nil {
				return result, fmt.Errorf("invalid catch all segment '%s' in HTTP rule: %s", segment, path)
			}

			result.Segments = append(result.Segments, Segment{
				Value: matches[1],
				Type:  SegmentTypeCatchAllSelector,
			})
		case strings.HasPrefix(segment, "{"):
			matches := selectorPattern.FindStringSubmatch(segment)
			if matches == nil {
				return result, fmt.Errorf("invalid selector segment '%s' in HTTP rule: %s", segment, path)
			}

			result.Segments = append(result.Segments, Segment{
				Value: matches[1],
				Type:  SegmentTypeSelector,
			})
		case segment == "*":
			result.Segments = append(result.Segments, Segment{
				Type: SegmentTypeWildcard,
			})
		default: // literal case.
			if !literalPattern.MatchString(segment) {
				return result, fmt.Errorf("invalid literal segment '%s' in HTTP rule: %s", segment, path)
			}

			result.Segments = append(result.Segments, Segment{
				Value: segment,
				Type:  SegmentTypeLiteral,
			})
		}
	}

	return result, nil
}
