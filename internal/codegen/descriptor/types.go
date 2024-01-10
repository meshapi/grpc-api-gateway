package descriptor

import (
	"fmt"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/internal/casing"
	"github.com/meshapi/grpc-rest-gateway/internal/httprule"
	"google.golang.org/protobuf/types/descriptorpb"
)

// GoPackage represents a golang package.
type GoPackage struct {
	// Path is the package path to the package.
	Path string
	// Name is the package name of the package
	Name string
	// Alias is an alias of the package unique within the current invocation of gRPC-Gateway generator.
	Alias string
}

// Standard returns whether the import is a golang standard package.
func (p GoPackage) Standard() bool {
	return !strings.Contains(p.Path, ".")
}

// String returns a string representation of this package in the form of import line in golang.
func (p GoPackage) String() string {
	if p.Alias == "" {
		return fmt.Sprintf("%q", p.Path)
	}
	return fmt.Sprintf("%s %q", p.Alias, p.Path)
}

// File wraps descriptorpb.FileDescriptorProto for richer features.
type File struct {
	*descriptorpb.FileDescriptorProto
	// GoPkg is the go package of the go file generated from this file.
	GoPkg GoPackage
	// GeneratedFilenamePrefix is used to construct filenames for generated
	// files associated with this source file.
	//
	// For example, the source file "dir/foo.proto" might have a filename prefix
	// of "dir/foo". Appending ".pb.go" produces an output file of "dir/foo.pb.go".
	GeneratedFilenamePrefix string
	// Messages is the list of messages defined in this file.
	Messages []*Message
	// Enums is the list of enums defined in this file.
	Enums []*Enum
	// Services is the list of services defined in this file.
	Services []*Service
}

// Pkg returns package name or alias if it's present
func (f *File) Pkg() string {
	pkg := f.GoPkg.Name
	if alias := f.GoPkg.Alias; alias != "" {
		pkg = alias
	}
	return pkg
}

// proto2 determines if the syntax of the file is proto2.
func (f *File) proto2() bool {
	return f.Syntax == nil || f.GetSyntax() == "proto2"
}

// Message describes a protocol buffer message types.
type Message struct {
	*descriptorpb.DescriptorProto
	// File is the file where the message is defined.
	File *File
	// Outers is a list of outer messages if this message is a nested type.
	Outers []string
	// Fields is a list of message fields.
	Fields []*Field
	// Index is proto path index of this message in File.
	Index int
	// ForcePrefixedName when set to true, prefixes a type with a package prefix.
	ForcePrefixedName bool
}

// FQMN returns a fully qualified message name of this message.
func (m *Message) FQMN() string {
	components := []string{""}
	if m.File.Package != nil {
		components = append(components, m.File.GetPackage())
	}
	components = append(components, m.Outers...)
	components = append(components, m.GetName())
	return strings.Join(components, ".")
}

// GoType returns a go type name for the message type.
// It prefixes the type name with the package alias if
// its belonging package is not "currentPackage".
func (m *Message) GoType(currentPackage string) string {
	var components []string
	components = append(components, m.Outers...)
	components = append(components, m.GetName())

	name := strings.Join(components, "_")
	if !m.ForcePrefixedName && m.File.GoPkg.Path == currentPackage {
		return name
	}
	return fmt.Sprintf("%s.%s", m.File.Pkg(), name)
}

// Enum describes a protocol buffer enum types.
type Enum struct {
	*descriptorpb.EnumDescriptorProto
	// File is the file where the enum is defined
	File *File
	// Outers is a list of outer messages if this enum is a nested type.
	Outers []string
	// Index is a enum index value.
	Index int
	// ForcePrefixedName when set to true, prefixes a type with a package prefix.
	ForcePrefixedName bool
}

// FQEN returns a fully qualified enum name of this enum.
func (e *Enum) FQEN() string {
	components := []string{""}
	if e.File.Package != nil {
		components = append(components, e.File.GetPackage())
	}
	components = append(components, e.Outers...)
	components = append(components, e.GetName())
	return strings.Join(components, ".")
}

// GoType returns a go type name for the enum type.
// It prefixes the type name with the package alias if
// its belonging package is not "currentPackage".
func (e *Enum) GoType(currentPackage string) string {
	var components []string
	components = append(components, e.Outers...)
	components = append(components, e.GetName())

	name := strings.Join(components, "_")
	if !e.ForcePrefixedName && e.File.GoPkg.Path == currentPackage {
		return name
	}
	return fmt.Sprintf("%s.%s", e.File.Pkg(), name)
}

// Service wraps descriptorpb.ServiceDescriptorProto for richer features.
type Service struct {
	*descriptorpb.ServiceDescriptorProto
	// File is the file where this service is defined.
	File *File
	// Methods is the list of methods defined in this service.
	Methods []*Method
	// ForcePrefixedName when set to true, prefixes a type with a package prefix.
	ForcePrefixedName bool
}

// FQSN returns the fully qualified service name of this service.
func (s *Service) FQSN() string {
	components := []string{""}
	if s.File.Package != nil {
		components = append(components, s.File.GetPackage())
	}
	components = append(components, s.GetName())
	return strings.Join(components, ".")
}

// InstanceName returns object name of the service with package prefix if needed
func (s *Service) InstanceName() string {
	if !s.ForcePrefixedName {
		return s.GetName()
	}
	return fmt.Sprintf("%s.%s", s.File.Pkg(), s.GetName())
}

// ClientConstructorName returns name of the Client constructor with package prefix if needed
func (s *Service) ClientConstructorName() string {
	constructor := "New" + s.GetName() + "Client"
	if !s.ForcePrefixedName {
		return constructor
	}
	return fmt.Sprintf("%s.%s", s.File.Pkg(), constructor)
}

// Parameter is a parameter provided in http requests
type Parameter struct {
	// FieldPath is a path to a proto field which this parameter is mapped to.
	FieldPath FieldPath
	// Target is the proto field which this parameter is mapped to.
	Target *Field
	// Method is the method which this parameter is used for.
	Method *Method
}

// Body describes a http (request|response) body to be sent to the (method|client).
// This is used in body and response_body options in google.api.HttpRule
type Body struct {
	// FieldPath is a path to a proto field which this parameter is mapped to.
	FieldPath FieldPath
}

// QueryParamAlias describes a query parameter alias, used to set/rename query params.
type QueryParamAlias struct {
	// Name is the name that will be read from the query parameters.
	Name string
	// FieldPath is a path to a proto field which this parameter is mapped to.
	FieldPath FieldPath
}

// QueryParameterCustomization describes the way query parameters are to get parsed.
type QueryParameterCustomization struct {
	// IgnoredFields are the field paths that are ignored.
	IgnoredFields []FieldPath
	// Aliases are the query parameter aliases.
	Aliases []QueryParamAlias
	// DisableAutoDiscovery disables auto discovery of query parameters and only allows the explicit declerations.
	DisableAutoDiscovery bool
}

// Binding describes how an HTTP endpoint is bound to a gRPC method.
type Binding struct {
	// Method is the method which the endpoint is bound to.
	Method *Method
	// Index is a zero-origin index of the binding in the target method
	Index int
	// PathTemplate is path template where this method is mapped to.
	PathTemplate httprule.Template
	// HTTPMethod is the HTTP method which this method is mapped to.
	HTTPMethod string
	// PathParameters is the list of parameters provided in HTTP request paths.
	PathParameters []Parameter
	// QueryParameterCustomization holds any customization for the way query parameters are handled.
	QueryParameterCustomization QueryParameterCustomization
	// Body describes parameters provided in HTTP request body.
	Body *Body
	// ResponseBody describes field in response struct to marshal in HTTP response body.
	ResponseBody *Body
}

// Method wraps descriptorpb.MethodDescriptorProto for richer features.
type Method struct {
	*descriptorpb.MethodDescriptorProto
	// Service is the service which this method belongs to.
	Service *Service
	// RequestType is the message type of requests to this method.
	RequestType *Message
	// ResponseType is the message type of responses from this method.
	ResponseType *Message
}

// FQMN returns a fully qualified rpc method name of this method.
func (m *Method) FQMN() string {
	var components []string
	components = append(components, m.Service.FQSN())
	components = append(components, m.GetName())
	return strings.Join(components, ".")
}

// Field wraps descriptorpb.FieldDescriptorProto for richer features.
type Field struct {
	*descriptorpb.FieldDescriptorProto
	// Message is the message type which this field belongs to.
	Message *Message
	// FieldMessage is the message type of the field.
	FieldMessage *Message
	// ForcePrefixedName when set to true, prefixes a type with a package prefix.
	ForcePrefixedName bool
}

// FQFN returns a fully qualified field name of this field.
func (f *Field) FQFN() string {
	return strings.Join([]string{f.Message.FQMN(), f.GetName()}, ".")
}

// FieldPath is a path to a field from a request message.
type FieldPath []FieldPathComponent

// String returns a string representation of the field path.
func (p FieldPath) String() string {
	components := make([]string, 0, len(p))
	for _, c := range p {
		components = append(components, c.Name)
	}
	return strings.Join(components, ".")
}

// IsNestedProto3 indicates whether the FieldPath is a nested Proto3 path.
func (p FieldPath) IsNestedProto3() bool {
	if len(p) > 1 && !p[0].Target.Message.File.proto2() {
		return true
	}
	return false
}

// IsOptionalProto3 indicates whether the FieldPath is a proto3 optional field.
func (p FieldPath) IsOptionalProto3() bool {
	if len(p) == 0 {
		return false
	}
	return p[0].Target.GetProto3Optional()
}

// AssignableExpr is an assignable expression in Go to be used to assign a value to the target field.
// It starts with "msgExpr", which is the go expression of the method request object. Before using
// such an expression the prep statements must be emitted first, in case the field path includes
// a oneof. See FieldPath.AssignableExprPrep.
func (p FieldPath) AssignableExpr(msgExpr string, currentPackage string) string {
	l := len(p)
	if l == 0 {
		return msgExpr
	}

	components := msgExpr
	for i, c := range p {
		// We need to check if the target is not proto3_optional first.
		// Under the hood, proto3_optional uses oneof to signal to old proto3 clients
		// that presence is tracked for this field. This oneof is known as a "synthetic" oneof.
		if !c.Target.GetProto3Optional() && c.Target.OneofIndex != nil {
			index := c.Target.OneofIndex
			msg := c.Target.Message
			oneOfName := casing.Camel(msg.GetOneofDecl()[*index].GetName())
			oneofFieldName := msg.GoType(currentPackage) + "_" + c.AssignableExpr()

			if c.Target.ForcePrefixedName {
				oneofFieldName = msg.File.Pkg() + "." + msg.GetName() + "_" + c.AssignableExpr()
			}

			components = components + "." + oneOfName + ".(*" + oneofFieldName + ")"
		}

		if i == l-1 {
			components = components + "." + c.AssignableExpr()
			continue
		}
		components = components + "." + c.ValueExpr()
	}
	return components
}

// AssignableExprPrep returns preparation statements for an assignable expression to assign a value
// to the target field. The Go expression of the method request object is "msgExpr". This is only
// needed for field paths that contain oneofs. Otherwise, an empty string is returned.
func (p FieldPath) AssignableExprPrep(msgExpr string, currentPackage string) string {
	l := len(p)
	if l == 0 {
		return ""
	}

	var preparations []string
	components := msgExpr
	for i, c := range p {
		// We need to check if the target is not proto3_optional first.
		// Under the hood, proto3_optional uses oneof to signal to old proto3 clients
		// that presence is tracked for this field. This oneof is known as a "synthetic" oneof.
		if !c.Target.GetProto3Optional() && c.Target.OneofIndex != nil {
			index := c.Target.OneofIndex
			msg := c.Target.Message
			oneOfName := casing.Camel(msg.GetOneofDecl()[*index].GetName())
			oneofFieldName := msg.GoType(currentPackage) + "_" + c.AssignableExpr()

			if c.Target.ForcePrefixedName {
				oneofFieldName = msg.File.Pkg() + "." + msg.GetName() + "_" + c.AssignableExpr()
			}

			components = components + "." + oneOfName
			s := `if %s == nil {
				%s =&%s{}
			} else if _, ok := %s.(*%s); !ok {
				return nil, metadata, status.Errorf(codes.InvalidArgument, "expect type: *%s, but: %%t\n",%s)
			}`

			preparations = append(preparations, fmt.Sprintf(s, components, components, oneofFieldName, components, oneofFieldName, oneofFieldName, components))
			components = components + ".(*" + oneofFieldName + ")"
		}

		if i == l-1 {
			components = components + "." + c.AssignableExpr()
			continue
		}
		components = components + "." + c.ValueExpr()
	}

	return strings.Join(preparations, "\n")
}

// FieldPathComponent is a path component in FieldPath
type FieldPathComponent struct {
	// Name is a name of the proto field which this component corresponds to.
	Name string
	// Target is the proto field which this component corresponds to.
	Target *Field
}

// AssignableExpr returns an assignable expression in go for this field.
func (c FieldPathComponent) AssignableExpr() string {
	return casing.Camel(c.Name)
}

// ValueExpr returns an expression in go for this field.
func (c FieldPathComponent) ValueExpr() string {
	if c.Target.Message.File.proto2() {
		return fmt.Sprintf("Get%s()", casing.Camel(c.Name))
	}
	return casing.Camel(c.Name)
}
