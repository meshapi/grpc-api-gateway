package genopenapi

import (
	"strings"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/fqn"
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

func (g *Generator) resolveMessageNames(messages []string) map[string]string {
	switch g.SchemaNamingStrategy {
	case SchemaNamingStrategyFQN:
		return resolveNamesFQN(messages)
	case SchemaNamingStrategySimple:
		return resolveNamesUniqueWithContext(messages, 0, ".")
	case SchemaNamingStrategySimpleWithVersion:
		return resolveNamesUniqueWithContext(messages, 1, ".")
	}

	return nil
}

func resolveNamesFQN(messages []string) map[string]string {
	result := make(map[string]string, len(messages))
	for _, message := range messages {
		result[message] = message[1:] // strip the leading dot here.
	}

	return result
}

// Take the names of every proto message and generates a unique reference by:
// first, separating each message name into its components by splitting at dots. Then,
// take the shortest suffix slice from each components slice that is unique among all
// messages, and convert it into a component name by taking extraContext additional
// components into consideration and joining all components with componentSeparator.
func resolveNamesUniqueWithContext(messages []string, extraContext int, componentSeparator string) map[string]string {
	packagesByDepth := make(map[int][]string)
	uniqueNames := make(map[string]string, len(messages))

	fqnItems := make([]fqn.Instance, len(messages))
	for index := range messages {
		fqnItems[index] = fqn.Parse(&messages[index])
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
		for ; depth < item.MaxDepth(); depth++ {
			if depth >= extraContext && count(packagesByDepth[depth], item.StringAtDepth(depth)) == 1 {
				break
			}
		}

		uniqueNames[messages[index]] = strings.Join(item.PartsAtDepth(depth), componentSeparator)
	}

	return uniqueNames
}
