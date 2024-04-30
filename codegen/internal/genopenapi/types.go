package genopenapi

import (
	"fmt"
	"strings"
)

type OutputMode uint8

const (
	OutputModePerService OutputMode = iota
	OutputModePerProtoFile
	OutputModeMerge
)

func (o OutputMode) String() string {
	switch o {
	case OutputModePerService:
		return "service"
	case OutputModePerProtoFile:
		return "proto"
	case OutputModeMerge:
		return "merge"
	default:
		return "<unknown>"
	}
}

func (o *OutputMode) Set(value string) error {
	switch strings.ToLower(value) {
	case "service":
		*o = OutputModePerService
	case "proto":
		*o = OutputModePerProtoFile
	case "merge":
		*o = OutputModeMerge
	default:
		return fmt.Errorf("unrecognized value: %q. Allowed values are 'service', 'proto' and 'merge'.", value)
	}

	return nil
}

type FieldNameMode uint8

const (
	FieldNameModeJSON FieldNameMode = iota
	FieldNameModeProto
)

func (f FieldNameMode) String() string {
	switch f {
	case FieldNameModeJSON:
		return "json"
	case FieldNameModeProto:
		return "proto"
	default:
		return "<unknown>"
	}
}

func (f *FieldNameMode) Set(value string) error {
	switch strings.ToLower(value) {
	case "json":
		*f = FieldNameModeJSON
	case "proto":
		*f = FieldNameModeProto
	default:
		return fmt.Errorf("unrecognized value: %q. Allowed values are 'json' and 'proto'", value)
	}

	return nil
}

type OutputFormat uint8

const (
	OutputFormatJSON OutputFormat = iota
	OutputFormatYAML
)

func (o OutputFormat) String() string {
	switch o {
	case OutputFormatJSON:
		return "json"
	case OutputFormatYAML:
		return "yaml"
	default:
		return "<unknown>"
	}
}

func (o *OutputFormat) Set(value string) error {
	switch strings.ToLower(value) {
	case "json":
		*o = OutputFormatJSON
	case "yaml":
		*o = OutputFormatYAML
	default:
		return fmt.Errorf("unrecognized value: %q. Allowed values are 'json' and 'yaml'.", value)
	}

	return nil
}

type SchemaNamingStrategy uint8

const (
	SchemaNamingStrategySimple SchemaNamingStrategy = iota
	SchemaNamingStrategyFQN
	SchemaNamingStrategySimpleWithVersion
)

func (s SchemaNamingStrategy) String() string {
	switch s {
	case SchemaNamingStrategySimple:
		return "simple"
	case SchemaNamingStrategyFQN:
		return "fqn"
	case SchemaNamingStrategySimpleWithVersion:
		return "simple+version"
	default:
		return "<unknown>"
	}
}

func (s *SchemaNamingStrategy) Set(value string) error {
	switch strings.ToLower(value) {
	case "simple":
		*s = SchemaNamingStrategySimple
	case "fqn":
		*s = SchemaNamingStrategyFQN
	case "simple+version":
		*s = SchemaNamingStrategySimpleWithVersion
	default:
		return fmt.Errorf("unrecognized value: %q. Allowed values are 'simple', 'version' and 'simple+version'.", value)
	}

	return nil
}

type TemplateArg struct {
	Key   string
	Value string
}

func (t TemplateArg) String() string {
	return fmt.Sprintf("%s = %s", t.Key, t.Value)
}

func (t *TemplateArg) Set(value string) error {
	index := strings.Index(value, "=")
	if index <= 0 || index+1 >= len(value)-1 {
		return fmt.Errorf("invalid format received for template argument %q, expected <key>=<value>", value)
	}

	t.Key = value[:index]
	t.Value = value[index+1:]
	return nil
}

type TemplateArgs []TemplateArg

func (t TemplateArgs) String() string {
	items := make([]string, len(t))
	for i, item := range t {
		items[i] = item.String()
	}
	return strings.Join(items, ",")
}

func (t *TemplateArgs) Set(value string) error {
	for _, item := range strings.Split(value, ",") {
		arg := TemplateArg{}
		if err := arg.Set(item); err != nil {
			return fmt.Errorf("failed to parse argument set: %w", err)
		}

		*t = append(*t, arg)
	}
	return nil
}

type SelectorSlice []string

func (s SelectorSlice) String() string {
	return strings.Join(s, ",")
}

func (s *SelectorSlice) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

type OperationIDMode uint8

const (
	OperationIDModeFQN OperationIDMode = iota
	OperationIDModeServiceAndMethod
	OperationIDModeMethod
)

func (o OperationIDMode) String() string {
	switch o {
	case OperationIDModeFQN:
		return "fqn"
	case OperationIDModeServiceAndMethod:
		return "service+method"
	case OperationIDModeMethod:
		return "method"
	default:
		return "n/a"
	}
}

func (o *OperationIDMode) Set(value string) error {
	switch strings.ToLower(value) {
	case "method":
		*o = OperationIDModeMethod
	case "service+method":
		*o = OperationIDModeServiceAndMethod
	case "fqn":
		*o = OperationIDModeFQN
	default:
		return fmt.Errorf("unrecognized value for operation id mode, expected 'simple' or 'fqn', got: %s", value)
	}

	return nil
}

type FieldRequiredMode uint8

const (
	FieldRequiredModeDisabled FieldRequiredMode = iota
	FieldRequiredModeRequireNonOptional
	FieldRequiredModeRequireNonOptionalScalar
)

func (f FieldRequiredMode) String() string {
	switch f {
	case FieldRequiredModeDisabled:
		return "disabled"
	case FieldRequiredModeRequireNonOptional:
		return "non_optional"
	case FieldRequiredModeRequireNonOptionalScalar:
		return "non_optional_scalar"
	default:
		return "n/a"
	}
}

func (f *FieldRequiredMode) Set(value string) error {
	switch strings.ToLower(value) {
	case "disabled":
		*f = FieldRequiredModeDisabled
	case "non_optional":
		*f = FieldRequiredModeRequireNonOptional
	case "non_optional_scalar":
		*f = FieldRequiredModeRequireNonOptionalScalar
	default:
		return fmt.Errorf("unrecognized value %q, expected 'disabled', 'non_optional' or 'non_optional_scalar'", value)
	}

	return nil
}

type FieldNullableMode uint8

const (
	FieldNullableModeDisabled FieldNullableMode = iota
	FieldNullableModeOptionalLabel
	FieldNullableModeNotRequired
)

func (f FieldNullableMode) String() string {
	switch f {
	case FieldNullableModeDisabled:
		return "disabled"
	case FieldNullableModeOptionalLabel:
		return "optional"
	case FieldNullableModeNotRequired:
		return "non_required"
	}

	return "n/a"
}

func (f *FieldNullableMode) Set(value string) error {
	switch strings.ToLower(value) {
	case "disabled":
		*f = FieldNullableModeDisabled
	case "optional":
		*f = FieldNullableModeOptionalLabel
	case "non_required":
		*f = FieldNullableModeNotRequired
	default:
		return fmt.Errorf("unrecognized value %q, expected 'disabled', 'optional' or 'non_required'", value)
	}

	return nil
}
