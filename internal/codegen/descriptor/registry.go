package descriptor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/meshapi/grpc-rest-gateway/internal/codegen/configpath"
	"github.com/meshapi/grpc-rest-gateway/internal/codegen/httpspec"
	"github.com/meshapi/grpc-rest-gateway/internal/codegen/plugin"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Registry is a registry of information extracted from pluginpb.CodeGeneratorRequest.
type Registry struct {
	// messages is a mapping from fully-qualified message name to descriptor
	messages map[string]*Message

	// enums is a mapping from fully-qualified enum name to descriptor
	enums map[string]*Enum

	// files is a mapping from file path to descriptor
	files map[string]*File

	// prefix is a prefix to be inserted to golang package paths generated from proto package names.
	prefix string

	// pkgMap is a user-specified mapping from file path to proto package.
	pkgMap map[string]string

	// pkgAliases is a mapping from package aliases to package paths in go which are already taken.
	pkgAliases map[string]string

	// httpSpecRegistry is HTTP specification registry which holds all the endpoint mappings.
	httpSpecRegistry *httpspec.Registry

	// PluginClient will be used (if specified) to find gateway config files.
	PluginClient *plugin.Client

	// GatewayFileLoadOptions holds gateway config file loading options.
	GatewayFileLoadOptions GatewayFileLoadOptions

	// SearchPath is the directory that is used to look for gateway configuration files.
	//
	// this search path can be relative or absolute, if relative, it will be from the current working directory.
	SearchPath string
}

// NewRegistry creates and initializes a new registry.
func NewRegistry() *Registry {
	return &Registry{
		messages:         map[string]*Message{},
		files:            map[string]*File{},
		enums:            map[string]*Enum{},
		pkgAliases:       map[string]string{},
		httpSpecRegistry: httpspec.NewRegistry(),
	}
}

// LoadFromPlugin loads all messages and enums in all files and services in all files that need generation.
func (r *Registry) LoadFromPlugin(gen *protogen.Plugin) error {
	return r.loadProtoFilesFromPlugin(gen)
}

func (r *Registry) loadProtoFilesFromPlugin(gen *protogen.Plugin) error {
	if r.GatewayFileLoadOptions.GlobalGatewayConfigFile != "" {
		if err := r.httpSpecRegistry.LoadFromFile(r.GatewayFileLoadOptions.GlobalGatewayConfigFile, ""); err != nil {
			return fmt.Errorf("failed to load global gateway config file: %w", err)
		}
	}

	filePaths := make([]string, 0, len(gen.FilesByPath))
	for filePath := range gen.FilesByPath {
		filePaths = append(filePaths, filePath)
	}
	sort.Strings(filePaths)

	for _, filePath := range filePaths {
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

	for _, filePath := range filePaths {
		if !gen.FilesByPath[filePath].Generate {
			continue
		}
		file := r.files[filePath]
		if err := r.loadServices(file); err != nil {
			return err
		}
	}

	writer := gen.NewGeneratedFile("output.txt", "")
	fmt.Fprintf(writer, "data: %+v", r.httpSpecRegistry)

	return nil
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
	for _, protoService := range file.GetService() {
		service := &Service{
			ServiceDescriptorProto: protoService,
			File:                   file,
			Methods:                []*Method{},
		}
		for _, protoMethod := range service.GetMethod() {
			if err := r.loadMethod(service, protoMethod); err != nil {
				return fmt.Errorf("failed to process method '%s': %w", protoMethod.GetName(), err)
			}
		}
	}

	return nil
}

func (r *Registry) loadMethod(service *Service, protoMethod *descriptorpb.MethodDescriptorProto) error {
	service.Methods = append(service.Methods, &Method{
		MethodDescriptorProto: &descriptorpb.MethodDescriptorProto{},
		Service:               service,
		RequestType:           &Message{},
		ResponseType:          &Message{},
	})

	return nil
}

func (r *Registry) loadMessagesInFile(file *File, outerPath []string, messages []*descriptorpb.DescriptorProto) {
	for index, protoMessage := range messages {
		message := &Message{
			DescriptorProto: protoMessage,
			File:            file,
			Outers:          outerPath,
			Index:           index,
		}
		for _, protoField := range protoMessage.GetField() {
			message.Fields = append(message.Fields, &Field{
				FieldDescriptorProto: protoField,
				Message:              message,
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
