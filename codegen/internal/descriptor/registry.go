package descriptor

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/configpath"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/httpspec"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/plugin"
	"github.com/meshapi/grpc-rest-gateway/dotpath"
	"github.com/meshapi/grpc-rest-gateway/pkg/httprule"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
)

// Registry is a registry of information extracted from pluginpb.CodeGeneratorRequest.
type Registry struct {
	// messages is a mapping from fully-qualified message name to descriptor
	messages map[string]*Message

	// enums is a mapping from fully-qualified enum name to descriptor
	enums map[string]*Enum

	// files is a mapping from file path to descriptor
	files map[string]*File

	// pkgAliases is a mapping from package aliases to package paths in go which are already taken.
	pkgAliases map[string]string

	// httpSpecRegistry is HTTP specification registry which holds all the endpoint mappings.
	httpSpecRegistry *httpspec.Registry

	RegistryOptions

	filePaths []string
}

// NewRegistry creates and initializes a new registry.
func NewRegistry(options RegistryOptions) *Registry {
	return &Registry{
		messages:         map[string]*Message{},
		files:            map[string]*File{},
		enums:            map[string]*Enum{},
		pkgAliases:       map[string]string{},
		httpSpecRegistry: httpspec.NewRegistry(),
		RegistryOptions:  options,
	}
}

// LoadFromPlugin loads all messages and enums in all files and services in all files that need generation.
func (r *Registry) LoadFromPlugin(gen *protogen.Plugin) error {
	return r.loadProtoFilesFromPlugin(gen)
}

func (r *Registry) loadProtoFilesFromPlugin(gen *protogen.Plugin) error {
	if r.GatewayFileLoadOptions.GlobalGatewayConfigFile != "" {
		filePath := filepath.Join(r.SearchPath, r.GatewayFileLoadOptions.GlobalGatewayConfigFile)
		if err := r.httpSpecRegistry.LoadFromFile(filePath, ""); err != nil {
			return fmt.Errorf("failed to load global gateway config file: %w", err)
		}
	}

	r.filePaths = make([]string, 0, len(gen.FilesByPath))
	for filePath := range gen.FilesByPath {
		r.filePaths = append(r.filePaths, filePath)
	}
	sort.Strings(r.filePaths)

	for _, filePath := range r.filePaths {
		r.loadIncludedFile(filePath, gen.FilesByPath[filePath])

		// if the file is a target of code genertaion, look for service mapping files.
		if protoFile := gen.FilesByPath[filePath]; protoFile.Generate {
			if err := r.loadEndpointsForFile(filePath, gen.FilesByPath[filePath]); err != nil {
				return err
			}

			protoPackage := protoFile.Proto.GetPackage()
			for _, protoService := range protoFile.Proto.GetService() {
				if err := r.httpSpecRegistry.LoadFromService(filePath, protoPackage, protoService); err != nil {
					return fmt.Errorf("failed to read embedded gateway configs from '%s': %w", filePath, err)
				}
			}
		}
	}

	for _, filePath := range r.filePaths {
		if !gen.FilesByPath[filePath].Generate {
			continue
		}

		file := r.files[filePath]
		if err := r.loadServices(file); err != nil {
			return err
		}
	}

	r.loadRPCStatus()
	return nil
}

// loadRPCStatus loads the rpc.Status object so that it can be found in the registry.
func (r *Registry) loadRPCStatus() {
	protoFiles := [2]*descriptorpb.FileDescriptorProto{
		protodesc.ToFileDescriptorProto((&anypb.Any{}).ProtoReflect().Descriptor().ParentFile()),
		protodesc.ToFileDescriptorProto((&statuspb.Status{}).ProtoReflect().Descriptor().ParentFile()),
	}

	for _, protoFile := range protoFiles {
		if _, exists := r.files[protoFile.GetName()]; exists {
			continue
		}

		f := &File{
			FileDescriptorProto: protoFile,
		}
		r.files[protoFile.GetName()] = f
		r.filePaths = append(r.filePaths, protoFile.GetName())
		r.loadMessagesInFile(f, nil, protoFile.MessageType)
		r.loadEnumsInFile(f, nil, protoFile.EnumType)
	}
}

func (r *Registry) loadEndpointsForFile(filePath string, protoFile *protogen.File) error {
	// only consider proto files that have service definitions.
	if len(protoFile.Services) == 0 {
		return nil
	}

	var configPath string

	// if plugin callback is available, use the plugin first.
	if r.PluginClient != nil && r.PluginClient.RegisteredCallbacks.Has(plugin.CallbackGatewayConfigFile) {
		path, err := r.PluginClient.GetGatewayConfigFile(context.Background(), protoFile.Proto)
		if err != nil {
			return fmt.Errorf("failed to get gateway config file from plugin: %w", err)
		}

		configPath = path
	}

	if configPath == "" {
		if r.GatewayFileLoadOptions.FilePattern == "" {
			return nil
		}

		path, err := configpath.Build(filePath, r.GatewayFileLoadOptions.FilePattern)
		if err != nil {
			return fmt.Errorf("failed to determine config file path: %w", err)
		}

		configPath = path
	}

	for _, ext := range [...]string{"yaml", "yml", "json"} {
		configFilePath := filepath.Join(r.SearchPath, configPath+"."+ext)

		if _, err := os.Stat(configFilePath); err != nil {
			if os.IsNotExist(err) {
				grpclog.Infof("looked for file %s, it was not found", configFilePath)
				continue
			}

			return fmt.Errorf("failed to stat file '%s': %w", configFilePath, err)
		}

		// file exists, try to load it.
		if err := r.httpSpecRegistry.LoadFromFile(configFilePath, protoFile.Proto.GetPackage()); err != nil {
			return fmt.Errorf("failed to load %s: %w", configFilePath, err)
		}

		return nil
	}

	return nil
}

func (r *Registry) loadIncludedFile(filePath string, protoFile *protogen.File) {
	pkg := GoPackage{
		Path: string(protoFile.GoImportPath),
		Name: string(protoFile.GoPackageName),
	}

	// if package cannot be reserved, keep iterating until it can.
	if !r.ReserveGoPackageAlias(pkg.Name, pkg.Path) {
		for i := 0; ; i++ {
			alias := fmt.Sprintf("%s_%d", pkg.Name, i)
			if r.ReserveGoPackageAlias(alias, pkg.Path) {
				pkg.Alias = alias
				break
			}
		}
	}
	f := &File{
		FileDescriptorProto:     protoFile.Proto,
		GoPkg:                   pkg,
		GeneratedFilenamePrefix: protoFile.GeneratedFilenamePrefix,
	}

	r.files[filePath] = f
	r.loadMessagesInFile(f, nil, protoFile.Proto.MessageType)
	r.loadEnumsInFile(f, nil, protoFile.Proto.EnumType)
}

func (r *Registry) loadServices(file *File) error {
	for index, protoService := range file.GetService() {
		service := &Service{
			ServiceDescriptorProto: protoService,
			File:                   file,
			Methods:                []*Method{},
			Index:                  index,
		}
		for index, protoMethod := range service.GetMethod() {
			fqmn := service.FQSN() + "." + protoMethod.GetName()
			binding, ok := r.httpSpecRegistry.LookupBinding(fqmn)
			if !ok {
				if r.GenerateUnboundMethods {
					// add default binding.
					binding = httpspec.EndpointSpec{
						Binding: &api.EndpointBinding{
							Selector: "",
							Pattern:  &api.EndpointBinding_Post{Post: fmt.Sprintf("/%s/%s", service.FQSN(), protoMethod.GetName())},
							Body:     "*",
						},
						SourceInfo: httpspec.SourceInfo{
							ProtoPackage: service.File.GetPackage(),
						},
					}
				} else {
					if r.WarnOnUnboundMethods {
						grpclog.Warningf("No HTTP binding specification found for method: %s.%s", service.GetName(), protoMethod.GetName())
					}
					continue
				}
			}

			if err := r.addMethodToService(service, index, protoMethod, binding); err != nil {
				return fmt.Errorf("failed to process method '%s': %w", protoMethod.GetName(), err)
			}
		}

		if len(service.Methods) == 0 {
			continue
		}

		file.Services = append(file.Services, service)
	}

	return nil
}

func (r *Registry) addMethodToService(
	service *Service, index int, protoMethod *descriptorpb.MethodDescriptorProto, binding httpspec.EndpointSpec) error {

	requestType, err := r.LookupMessage(service.File.GetPackage(), protoMethod.GetInputType())
	if err != nil {
		return err
	}

	responseType, err := r.LookupMessage(service.File.GetPackage(), protoMethod.GetOutputType())
	if err != nil {
		return err
	}

	method := &Method{
		MethodDescriptorProto: protoMethod,
		Service:               service,
		RequestType:           requestType,
		ResponseType:          responseType,
		Index:                 index,
	}

	method.Bindings, err = r.mapBindings(method, binding)
	if err != nil {
		return err
	}

	service.Methods = append(service.Methods, method)
	return nil
}

func (r *Registry) mapBindings(md *Method, spec httpspec.EndpointSpec) ([]*Binding, error) {
	var bindings []*Binding

	type BindingInput struct {
		Method                          string
		Path                            string
		Body                            string
		ResponseBody                    string
		Index                           int
		DisableQueryParamsAutoDiscovery bool
		QueryParams                     []*api.QueryParameterBinding
		StreamConfig                    *api.StreamConfig
	}

	insertBinding := func(input BindingInput) error {
		tpl, err := httprule.Parse(input.Path)
		if err != nil {
			return fmt.Errorf("failed to parse HTTP rule %s: %w (%s)", input.Path, err, spec.SourceInfo.Filename)
		}

		if md.GetClientStreaming() && tpl.HasVariables() {
			return fmt.Errorf("cannot use path parameters in client streaming: %s", md.FQMN())
		}

		binding := Binding{
			Method:       md,
			Index:        input.Index,
			PathTemplate: tpl,
			HTTPMethod:   input.Method,
		}

		for _, segment := range tpl.Segments {
			if segment.Type == httprule.SegmentTypeCatchAllSelector || segment.Type == httprule.SegmentTypeSelector {
				param, err := r.mapParam(md, segment.Value)
				if err != nil {
					return fmt.Errorf("failed to map path parameter in %s: %w", input.Path, err)
				}

				binding.PathParameters = append(binding.PathParameters, param)
			}
		}

		binding.Body, err = r.mapBody(md, input.Body)
		if err != nil {
			return fmt.Errorf("failed to parse request body selector %q: %w", input.Body, err)
		}

		binding.ResponseBody, err = r.mapResponseBody(md, input.ResponseBody)
		if err != nil {
			return fmt.Errorf("failed to parse response body selector %q: %w", input.ResponseBody, err)
		}

		binding.QueryParameterCustomization.DisableAutoDiscovery = input.DisableQueryParamsAutoDiscovery

		queryParamFilter := binding.QueryParameterFilter()
		for _, queryParam := range input.QueryParams {
			fields, err := r.resolveFieldPath(md.RequestType, queryParam.GetSelector(), false)
			if err != nil {
				return fmt.Errorf("failed to resolve field at selector %q: %w", queryParam.GetSelector(), err)
			}

			// if query param is already used by another target, error out.
			alreadyBound := (binding.Body != nil && len(binding.Body.FieldPath) == 0) ||
				queryParamFilter.HasCommonPrefix(dotpath.Parse(&queryParam.Selector))
			if alreadyBound {
				return fmt.Errorf(
					"cannot use selector %q for query parameter %q because it will already be read from payload/path params",
					queryParam.Selector, queryParam.Name)
			}

			if !fields[len(fields)-1].Target.IsScalarType() {
				return fmt.Errorf(
					"cannot use selector %q for query parameter %q because it points to a"+
						" Protobuf message type, only scalar types can be used",
					queryParam.Selector, queryParam.Name)
			}

			if queryParam.Ignore {
				binding.QueryParameterCustomization.IgnoredFields = append(
					binding.QueryParameterCustomization.IgnoredFields,
					FieldPath(fields))
			} else {
				name := queryParam.Name
				customName := true
				if name == "" {
					name = FieldPath(fields).String()
					customName = false
				}

				binding.QueryParameterCustomization.Aliases = append(
					binding.QueryParameterCustomization.Aliases,
					QueryParamAlias{Name: name, FieldPath: FieldPath(fields), CustomName: customName})
			}
		}

		binding.QueryParameters, err = buildQueryParameters(&binding, r)
		if err != nil {
			return err
		}

		if binding.Method.GetClientStreaming() || binding.Method.GetServerStreaming() {
			if input.StreamConfig != nil {
				binding.StreamConfig = StreamConfig{
					AllowWebsocket:       !input.StreamConfig.DisableWebsockets,
					AllowSSE:             !input.StreamConfig.DisableSse,
					AllowChunkedTransfer: !input.StreamConfig.DisableChunkedTransfer,
				}
			} else {
				binding.StreamConfig = StreamConfig{
					AllowWebsocket:       true,
					AllowSSE:             true,
					AllowChunkedTransfer: true,
				}
			}

			if binding.Method.GetServerStreaming() && !binding.HasAnyStreamingMethod() {
				return fmt.Errorf(
					"streaming method %q does not support any streaming method (sse, websocket, chunked transfer),"+
						" note that you must use GET for SSE and Websocket streaming methods", md.FQMN())
			}
		}

		bindings = append(bindings, &binding)
		return nil
	}

	method, path, err := parseEndpointPattern(spec.Binding)
	if err != nil {
		return nil, fmt.Errorf("failed to process binding for '%s' (%s): %w", md.FQMN(), spec.SourceInfo.Filename, err)
	}

	input := BindingInput{
		Method:                          method,
		Path:                            path,
		Body:                            spec.Binding.Body,
		ResponseBody:                    spec.Binding.ResponseBody,
		Index:                           0,
		DisableQueryParamsAutoDiscovery: spec.Binding.DisableQueryParamDiscovery,
		QueryParams:                     spec.Binding.GetQueryParams(),
		StreamConfig:                    spec.Binding.Stream,
	}

	if err := insertBinding(input); err != nil {
		return nil, err
	}

	for index, additionalBinding := range spec.Binding.GetAdditionalBindings() {
		method, path, err = parseAdditionalEndpointPattern(additionalBinding)
		if err != nil {
			return nil, fmt.Errorf("failed to process binding for '%s' (%s): %w", md.FQMN(), spec.SourceInfo.Filename, err)
		}

		input = BindingInput{
			Method:                          method,
			Path:                            path,
			Body:                            additionalBinding.Body,
			ResponseBody:                    additionalBinding.ResponseBody,
			Index:                           index + 1,
			DisableQueryParamsAutoDiscovery: additionalBinding.DisableQueryParamDiscovery,
			QueryParams:                     additionalBinding.GetQueryParams(),
			StreamConfig:                    additionalBinding.Stream,
		}

		if err := insertBinding(input); err != nil {
			return nil, err
		}
	}

	return bindings, nil
}

// buildQueryParameters builds a list of all bound query parameters.
func buildQueryParameters(b *Binding, registry *Registry) ([]QueryParameter, error) {
	queryFilter := b.QueryParameterFilter()

	// if body is not defined or it is '*' (everything), there are no query parameters.
	if b.Body != nil && len(b.Body.FieldPath) == 0 {
		return nil, nil
	}

	var queryParams []QueryParameter

	for _, alias := range b.QueryParameterCustomization.Aliases {
		field := alias.FieldPath[len(alias.FieldPath)-1].Target

		repeated := field.HasRepeatedLabel()
		if !field.IsScalarType() {
			if repeated {
				return nil, fmt.Errorf(
					"invalid query param %q: only primitive and enum types can be used with a repeated label in query params",
					alias.Name)
			}

			message, err := registry.LookupMessage(field.Message.FQMN(), field.GetTypeName())
			if err != nil {
				return nil, fmt.Errorf("failed to find message %q: %w", field.GetTypeName(), err)
			}
			// if the map entry's value type is not scalar, it cannot be parsed in query parameters.
			if message.IsMapEntry() && !message.Fields[1].IsScalarType() {
				return nil, fmt.Errorf(
					"invalid query param %q: only primitive and enum types can be used in map values", alias.Name)
			}
		}

		queryParams = append(queryParams, QueryParameter{
			Name:        alias.Name,
			FieldPath:   alias.FieldPath,
			NameIsAlias: alias.CustomName,
		})
	}

	if b.QueryParameterCustomization.DisableAutoDiscovery {
		return queryParams, nil
	}

	// Item is the queue item for processing messages that have query parameters.
	type Item struct {
		Message   *Message
		Prefix    string
		FieldPath FieldPath
	}

	queue := []Item{
		{Message: b.Method.RequestType},
	}

	for index := 0; index < len(queue); index++ {
		item := queue[index]
		for _, field := range item.Message.Fields {
			var name string
			if item.Prefix != "" {
				name = item.Prefix + "." + field.GetName()
			} else {
				name = field.GetName()
			}
			if queryFilter.HasCommonPrefixString(name) {
				continue
			}

			repeated := field.HasRepeatedLabel()

			if !field.IsScalarType() {
				// go deep inside.
				message, err := registry.LookupMessage(item.Message.FQMN(), field.GetTypeName())
				if err != nil {
					return nil, fmt.Errorf("failed to look up nested message type %q: %s", field.GetTypeName(), err)
				}

				// if message is a map entry, treat it as a scalar type.
				if !message.IsMapEntry() {
					if repeated {
						continue
					}

					queue = append(queue, Item{
						Message:   message,
						Prefix:    name,
						FieldPath: append(item.FieldPath, FieldPathComponent{Name: field.GetName(), Target: field}),
					})
					continue
				} else if !message.Fields[1].IsScalarType() {
					continue
				}
			}

			queryParams = append(queryParams, QueryParameter{
				Name:      name,
				FieldPath: append(item.FieldPath, FieldPathComponent{Name: field.GetName(), Target: field}),
			})
		}
	}

	return queryParams, nil
}

func (r *Registry) mapBody(md *Method, path string) (*Body, error) {
	switch path {
	case "":
		return nil, nil
	case "*":
		return &Body{FieldPath: nil}, nil
	}

	msg := md.RequestType
	fields, err := r.resolveFieldPath(msg, path, false)
	if err != nil {
		return nil, err
	}

	return &Body{FieldPath: FieldPath(fields)}, nil
}

func (r *Registry) mapResponseBody(md *Method, path string) (*Body, error) {
	msg := md.ResponseType
	switch path {
	case "", "*":
		return nil, nil
	}
	fields, err := r.resolveFieldPath(msg, path, false)
	if err != nil {
		return nil, err
	}
	return &Body{FieldPath: FieldPath(fields)}, nil
}

func (r *Registry) mapParam(md *Method, path string) (Parameter, error) {
	msg := md.RequestType
	fields, err := r.resolveFieldPath(msg, path, true)
	if err != nil {
		return Parameter{}, err
	}
	l := len(fields)
	if l == 0 {
		return Parameter{}, fmt.Errorf("invalid field access list for %s", path)
	}
	target := fields[l-1].Target
	if !target.IsScalarType() {
		return Parameter{}, fmt.Errorf(
			"%s.%s: %s is a protobuf message type. Protobuf message types cannot be used as path parameters,"+
				" use a scalar value type (such as string) instead",
			md.Service.GetName(), md.GetName(), path)
	}

	return Parameter{
		FieldPath: FieldPath(fields),
		Method:    md,
		Target:    fields[l-1].Target,
	}, nil
}

// lookupField looks up a field named "name" within "msg".
// It returns nil if no such field found.
func lookupField(msg *Message, name string) *Field {
	for _, field := range msg.Fields {
		if field.GetName() == name {
			return field
		}
	}

	return nil
}

// resolveFieldPath resolves "path" into a list of fieldDescriptor, starting from "msg".
func (r *Registry) resolveFieldPath(msg *Message, path string, isPathParam bool) ([]FieldPathComponent, error) {
	if path == "" {
		return nil, nil
	}

	root := msg
	var result []FieldPathComponent
	for i, c := range strings.Split(path, ".") {
		if i > 0 {
			f := result[i-1].Target
			switch f.GetType() {
			case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE, descriptorpb.FieldDescriptorProto_TYPE_GROUP:
				var err error
				msg, err = r.LookupMessage(msg.FQMN(), f.GetTypeName())
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("not an aggregate type: %s in %s", f.GetName(), path)
			}
		}

		field := lookupField(msg, c)
		if field == nil {
			return nil, fmt.Errorf("no field %q found in %s", path, root.GetName())
		}

		if isPathParam && field.GetProto3Optional() {
			return nil, fmt.Errorf("optional field not allowed in field path: %s in %s", field.GetName(), path)
		}

		result = append(result, FieldPathComponent{Name: c, Target: field})
	}

	return result, nil
}

// parseEndpointPattern returns HTTP method and path.
func parseEndpointPattern(spec *api.EndpointBinding) (string, string, error) {
	if spec.Pattern == nil {
		return "", "", fmt.Errorf("No pattern specified in HTTP rule")
	}

	switch pattern := spec.Pattern.(type) {
	case *api.EndpointBinding_Custom:
		return strings.ToUpper(pattern.Custom.Method), pattern.Custom.Path, nil
	case *api.EndpointBinding_Get:
		return http.MethodGet, pattern.Get, nil
	case *api.EndpointBinding_Patch:
		return http.MethodPatch, pattern.Patch, nil
	case *api.EndpointBinding_Post:
		return http.MethodPost, pattern.Post, nil
	case *api.EndpointBinding_Put:
		return http.MethodPut, pattern.Put, nil
	case *api.EndpointBinding_Delete:
		return http.MethodDelete, pattern.Delete, nil
	default:
		return "", "", fmt.Errorf("No pattern specified in HTTP rule")
	}
}

// parseAdditionalEndpointPattern returns HTTP method and path.
func parseAdditionalEndpointPattern(spec *api.AdditionalEndpointBinding) (string, string, error) {
	if spec.Pattern == nil {
		return "", "", fmt.Errorf("No pattern specified in HTTP rule")
	}

	switch pattern := spec.Pattern.(type) {
	case *api.AdditionalEndpointBinding_Custom:
		return strings.ToUpper(pattern.Custom.Method), pattern.Custom.Path, nil
	case *api.AdditionalEndpointBinding_Get:
		return http.MethodGet, pattern.Get, nil
	case *api.AdditionalEndpointBinding_Patch:
		return http.MethodPatch, pattern.Patch, nil
	case *api.AdditionalEndpointBinding_Post:
		return http.MethodPost, pattern.Post, nil
	case *api.AdditionalEndpointBinding_Put:
		return http.MethodPut, pattern.Put, nil
	case *api.AdditionalEndpointBinding_Delete:
		return http.MethodDelete, pattern.Delete, nil
	default:
		return "", "", fmt.Errorf("No pattern specified in HTTP rule")
	}
}

func (r *Registry) loadMessagesInFile(file *File, outerPath []string, messages []*descriptorpb.DescriptorProto) {
	for index, protoMessage := range messages {
		message := &Message{
			DescriptorProto: protoMessage,
			File:            file,
			Outers:          outerPath,
			Index:           index,
		}
		for index, protoField := range protoMessage.GetField() {
			message.Fields = append(message.Fields, &Field{
				FieldDescriptorProto: protoField,
				Message:              message,
				Index:                index,
			})
		}

		file.Messages = append(file.Messages, message)
		r.messages[message.FQMN()] = message

		var outers []string
		outers = append(outers, outerPath...)
		outers = append(outers, message.GetName())
		r.loadMessagesInFile(file, outers, message.GetNestedType())
		r.loadEnumsInFile(file, outers, message.GetEnumType())
	}
}

func (r *Registry) loadEnumsInFile(file *File, outerPath []string, enums []*descriptorpb.EnumDescriptorProto) {
	for index, protoEnum := range enums {
		enum := &Enum{
			EnumDescriptorProto: protoEnum,
			File:                file,
			Outers:              outerPath,
			Index:               index,
		}
		file.Enums = append(file.Enums, enum)
		r.enums[enum.FQEN()] = enum
	}
}

// ReserveGoPackageAlias reserves the unique alias of go package.
// If succeeded, the alias will be never used for other packages in generated go files.
// If failed, the alias is already taken by another package, so you need to use another
// alias for the package in your go files.
func (r *Registry) ReserveGoPackageAlias(alias, pkgPath string) bool {
	if taken, ok := r.pkgAliases[alias]; ok {
		return taken == pkgPath
	}

	r.pkgAliases[alias] = pkgPath
	return true
}

// LookupMessage looks up a message type by "name".
// It tries to resolve "name" from "location" if "name" is a relative message name.
//
// location must be a dot separated proto package to build a FQMN.
func (r *Registry) LookupMessage(location, name string) (*Message, error) {
	// If a message starts with a dot, it indicates that it is an absolute name.
	if strings.HasPrefix(name, ".") {
		m, ok := r.messages[name]
		if !ok {
			return nil, fmt.Errorf("no message found: %s", name)
		}

		return m, nil
	}

	if !strings.HasPrefix(location, ".") {
		location = "." + location
	}

	locationPath := dotpath.Parse(&location)
	for i := 0; i < locationPath.NumberOfSegments(); i++ {
		prefix := locationPath.TrimmedSuffix(i)
		if m, ok := r.messages[prefix+"."+name]; ok {
			return m, nil
		}
	}

	return nil, fmt.Errorf("no message found: %s", name)
}

// LookupFile looks up a file by name.
func (r *Registry) LookupFile(name string) (*File, error) {
	file, ok := r.files[name]
	if !ok {
		return nil, fmt.Errorf("no such file given: %s", name)
	}

	return file, nil
}

// LookupEnum looks up a enum type by "name".
// It tries to resolve "name" from "location" if "name" is a relative enum name.
func (r *Registry) LookupEnum(location, name string) (*Enum, error) {
	if strings.HasPrefix(name, ".") {
		e, ok := r.enums[name]
		if !ok {
			return nil, fmt.Errorf("no enum found: %s", name)
		}

		return e, nil
	}

	if !strings.HasPrefix(location, ".") {
		location = "." + location
	}

	locationPath := dotpath.Parse(&location)
	for i := 0; i < locationPath.NumberOfSegments(); i++ {
		prefix := locationPath.TrimmedSuffix(i)
		if e, ok := r.enums[prefix+"."+name]; ok {
			return e, nil
		}
	}

	return nil, fmt.Errorf("no enum found: %s", name)
}

// UnboundExternalHTTPSpecs returns the list of external HTTP specs
// which do not have a matching method in the registry.
func (r *Registry) UnboundExternalHTTPSpecs() []httpspec.EndpointSpec {
	allServiceMethods := map[string]struct{}{}
	for _, file := range r.files {
		for _, service := range file.Services {
			fqsn := service.FQSN()
			for _, method := range service.ServiceDescriptorProto.GetMethod() {
				method := fqsn + "." + method.GetName()
				allServiceMethods[method] = struct{}{}
			}
		}
	}

	var missingMethods []httpspec.EndpointSpec
	r.httpSpecRegistry.Iterate(func(fqmn string, spec httpspec.EndpointSpec) {
		if _, ok := allServiceMethods[fqmn]; !ok {
			missingMethods = append(missingMethods, spec)
		}
	})

	return missingMethods
}

// Iterate iterates over all processed files.
func (r *Registry) Iterate(cb func(fielPath string, protoFile *File) error) error {
	for _, filePath := range r.filePaths {
		if err := cb(filePath, r.files[filePath]); err != nil {
			return err
		}
	}

	return nil
}
