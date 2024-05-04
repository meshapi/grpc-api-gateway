package protocomment

import (
	"github.com/meshapi/grpc-api-gateway/codegen/internal/descriptor"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	// TypeMessage is the number of the message type index in the FileDescriptorProto.
	TypeMessage       = 4
	TypeNestedType    = 3
	TypeNestedEnum    = 4
	TypeEnum          = 5
	TypeEnumValue     = 2
	TypeMessageField  = 2
	TypeService       = 6
	TypeServiceMethod = 2
)

type Location = descriptorpb.SourceCodeInfo_Location

// File is an indexed data structure for comments.
type File struct {
	Messages map[int32]*Message
	Enums    map[int32]*Enum
	Services map[int32]*Service
}

type Enum struct {
	*Location

	Values map[int32]*Location
}

type Message struct {
	*Location

	NestedTypes map[int32]*Message
	NestedEnums map[int32]*Enum
	Fields      map[int32]*Location
}

type Service struct {
	*Location

	Methods map[int32]*Location
}

// Registry holds indexed files that hold comments and locations for various proto types for quick lookups.
type Registry struct {
	descriptorRegistry *descriptor.Registry
	files              map[*descriptor.File]*File
}

func NewRegistry(descriptorRegistry *descriptor.Registry) *Registry {
	return &Registry{
		descriptorRegistry: descriptorRegistry,
		files:              make(map[*descriptor.File]*File),
	}
}

func (r *Registry) processMessagePath(index int, message *Message, location *descriptorpb.SourceCodeInfo_Location) {
	if index >= len(location.Path) {
		message.Location = location
		return
	}

	switch location.Path[index] {
	case TypeNestedType:
		if message.NestedTypes == nil {
			message.NestedTypes = make(map[int32]*Message)
		}
		nestedMessage, ok := message.NestedTypes[location.Path[index+1]]
		if !ok {
			nestedMessage = &Message{}
			message.NestedTypes[location.Path[index+1]] = nestedMessage
		}
		r.processMessagePath(index+2, nestedMessage, location)
	case TypeNestedEnum:
		if message.NestedEnums == nil {
			message.NestedEnums = make(map[int32]*Enum)
		}
		nestedEnum, ok := message.NestedEnums[location.Path[index+1]]
		if !ok {
			nestedEnum = &Enum{}
			message.NestedEnums[location.Path[index+1]] = nestedEnum
		}
		r.processEnumPath(index+2, nestedEnum, location)
	case TypeMessageField:
		if index+2 < len(location.Path) {
			return
		}
		if message.Fields == nil {
			message.Fields = make(map[int32]*Location)
		}
		message.Fields[location.Path[index+1]] = location
	}
}

func (r *Registry) processEnumPath(index int, enum *Enum, location *descriptorpb.SourceCodeInfo_Location) {
	if index >= len(location.Path) {
		enum.Location = location
		return
	}

	switch location.Path[index] {
	case TypeEnumValue:
		if index+2 < len(location.Path) {
			return
		}
		if enum.Values == nil {
			enum.Values = make(map[int32]*Location)
		}
		enum.Values[location.Path[index+1]] = location
	}
}

func (r *Registry) processServicePath(index int, service *Service, location *descriptorpb.SourceCodeInfo_Location) {
	if index >= len(location.Path) {
		service.Location = location
		return
	}

	switch location.Path[index] {
	case TypeServiceMethod:
		if index+2 < len(location.Path) {
			return
		}
		if service.Methods == nil {
			service.Methods = make(map[int32]*Location)
		}
		service.Methods[location.Path[index+1]] = location
	}
}

func (r *Registry) evaluateOrGetFile(file *descriptor.File) *File {
	if indexedFile, ok := r.files[file]; ok {
		return indexedFile
	}

	indexedFile := &File{}

	for _, sci := range file.GetSourceCodeInfo().GetLocation() {
		if len(sci.Path) == 0 {
			continue
		}

		if sci.GetLeadingComments() == "" && sci.GetTrailingComments() == "" {
			continue
		}

		switch sci.Path[0] {
		case TypeMessage:
			if len(sci.Path)%2 != 0 {
				continue
			}

			if indexedFile.Messages == nil {
				indexedFile.Messages = make(map[int32]*Message)
			}

			message, ok := indexedFile.Messages[sci.Path[1]]
			if !ok {
				message = &Message{}
				indexedFile.Messages[sci.Path[1]] = message
			}

			r.processMessagePath(2, message, sci)
		case TypeEnum:
			if len(sci.Path)%2 != 0 {
				continue
			}

			if indexedFile.Enums == nil {
				indexedFile.Enums = make(map[int32]*Enum)
			}

			enum, ok := indexedFile.Enums[sci.Path[1]]
			if !ok {
				enum = &Enum{}
				indexedFile.Enums[sci.Path[1]] = enum
			}

			r.processEnumPath(2, enum, sci)
		case TypeService:
			if len(sci.Path)%2 != 0 {
				continue
			}

			if indexedFile.Services == nil {
				indexedFile.Services = make(map[int32]*Service)
			}

			service, ok := indexedFile.Services[sci.Path[1]]
			if !ok {
				service = &Service{}
				indexedFile.Services[sci.Path[1]] = service
			}

			r.processServicePath(2, service, sci)
		}
	}

	r.files[file] = indexedFile
	return indexedFile
}

func (r *Registry) resolveOuters(pkg string, file *File, outers []string) *Message {
	var cursor *Message
	root := "." + pkg

	for _, name := range outers {
		fqmn := root + "." + name
		msg, err := r.descriptorRegistry.LookupMessage("", fqmn)
		if err != nil {
			return nil
		}
		root = fqmn
		if cursor == nil {
			item, ok := file.Messages[int32(msg.Index)]
			if !ok {
				return nil
			}
			cursor = item
		} else {
			item, ok := cursor.NestedTypes[int32(msg.Index)]
			if !ok {
				return nil
			}
			cursor = item
		}
	}

	return cursor
}

func (r *Registry) LookupMessage(message *descriptor.Message) *Message {
	file := r.evaluateOrGetFile(message.File)
	if file == nil {
		return nil
	}

	if len(message.Outers) > 0 {
		cursor := r.resolveOuters(message.File.GetPackage(), file, message.Outers)
		if cursor == nil {
			return nil
		}

		result, ok := cursor.NestedTypes[int32(message.Index)]
		if !ok {
			return nil
		}

		return result
	}

	return file.Messages[int32(message.Index)]
}

// LookupField looks up comments for a field.
//
// NOTE: Avoid using this method unless an arbitrary field access is needed.
// A more efficient way when iterating over fields of a message is to use the Message object
// directly.
func (r *Registry) LookupField(field *descriptor.Field) *Location {
	comments := r.LookupMessage(field.Message)
	if comments == nil || comments.Fields == nil {
		return nil
	}

	return comments.Fields[int32(field.Index)]
}

func (r *Registry) LookupEnum(enum *descriptor.Enum) *Enum {
	file := r.evaluateOrGetFile(enum.File)
	if file == nil {
		return nil
	}

	if len(enum.Outers) > 0 {
		cursor := r.resolveOuters(enum.File.GetPackage(), file, enum.Outers)
		if cursor == nil {
			return nil
		}

		result, ok := cursor.NestedEnums[int32(enum.Index)]
		if !ok {
			return nil
		}

		return result
	}

	return file.Enums[int32(enum.Index)]
}

func (r *Registry) LookupService(service *descriptor.Service) *Service {
	file := r.evaluateOrGetFile(service.File)
	if file == nil {
		return nil
	}

	return file.Services[int32(service.Index)]
}
