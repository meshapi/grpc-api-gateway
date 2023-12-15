package descriptor

import (
	"fmt"
	"sort"

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
}

// NewRegistry creates and initializes a new registry.
func NewRegistry() *Registry {
	return &Registry{
		messages:   map[string]*Message{},
		files:      map[string]*File{},
		pkgAliases: map[string]string{},
	}
}

// LoadFromPlugin loads all messages and enums in all files and services in all files that need generation.
func (r *Registry) LoadFromPlugin(gen *protogen.Plugin) error {
	return r.loadProtoFilesFromPlugin(gen)
}

func (r *Registry) loadProtoFilesFromPlugin(gen *protogen.Plugin) error {
	filePaths := make([]string, 0, len(gen.FilesByPath))
	for filePath := range gen.FilesByPath {
		filePaths = append(filePaths, filePath)
	}
	sort.Strings(filePaths)

	for _, filePath := range filePaths {
		r.loadIncludedFile(filePath, gen.FilesByPath[filePath])
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
	//r.registerEnum(f, nil, protoFile.Proto.EnumType)
}

func (r *Registry) loadServices(file *File) error {
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
		//r.registerEnum(file, outers, m.GetEnumType())
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

func (r *Registry) VisitMessages(cb func(*Message)) {
	for _, message := range r.messages {
		cb(message)
	}
}
