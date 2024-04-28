package genopenapi

import (
	"strings"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/internal"
	"github.com/meshapi/grpc-rest-gateway/dotpath"
)

func (g *Generator) fieldName(field *descriptor.Field) string {
	switch g.FieldNameMode {
	case FieldNameModeJSON:
		return field.GetJsonName()
	case FieldNameModeProto:
		return field.GetName()
	}
	panic("unsupported field type " + g.FieldNameMode.String() + " received")
}

func (g *Generator) resolveTypeNames(protoTypes []internal.ProtoTypeName) map[string]string {
	switch g.SchemaNamingStrategy {
	case SchemaNamingStrategyFQN:
		return resolveNamesFQN(protoTypes)
	case SchemaNamingStrategySimple:
		return resolveNamesUniqueWithContext(protoTypes, 0, ".")
	case SchemaNamingStrategySimpleWithVersion:
		return resolveNamesUniqueWithContext(protoTypes, 1, ".")
	}

	return nil
}

func resolveNamesFQN(types []internal.ProtoTypeName) map[string]string {
	result := make(map[string]string, len(types))
	for _, protoType := range types {
		result[protoType.FQN] = protoType.FQN[1:] // strip the leading dot here.
	}

	return result
}

// Take the names of every proto message and generates a unique reference by:
// first, separating each message name into its components by splitting at dots. Then,
// take the shortest suffix slice from each components slice that is unique among all
// messages, and convert it into a component name by taking extraContext additional
// components into consideration and joining all components with componentSeparator.
func resolveNamesUniqueWithContext(types []internal.ProtoTypeName, extraContext int, componentSeparator string) map[string]string {
	packagesByDepth := make(map[int][]string)
	uniqueNames := make(map[string]string, len(types))

	fqnItems := make([]dotpath.Instance, len(types))
	for index := range types {
		fqnItems[index] = dotpath.Parse(&types[index].FQN)
	}

	for _, item := range fqnItems {
		for depth := 0; depth < item.MaxDepth(); depth++ {
			packagesByDepth[depth] = append(packagesByDepth[depth], item.StringAtDepth(depth))
		}
	}

	count := func(list []string, item string) int {
		i := 0
		for _, element := range list {
			if element == item {
				i++
			}
		}
		return i
	}

	for index, item := range fqnItems {
		depth := 0
		desiredContext := extraContext + int(types[index].OuterLength)
		for ; depth < item.MaxDepth(); depth++ {
			if depth >= desiredContext && count(packagesByDepth[depth], item.StringAtDepth(depth)) == 1 {
				break
			}
		}

		uniqueNames[types[index].FQN] = strings.Join(item.PartsAtDepth(depth), componentSeparator)
	}

	return uniqueNames
}
