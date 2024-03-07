package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

type SourceInfo struct {
	Filename     string
	ProtoPackage string
	ProtoFile    *descriptor.File
}

func (s *SourceInfo) IsGlobalConfigFile() bool {
	return s.ProtoFile == nil
}

// Registry contains references to all configuration files.
type Registry struct {
	// RootDocument is the global and top-level document loaded from the global config.
	RootDocument *Document

	documents map[*descriptor.File]*Document

	// schemas map[string]string
}

func NewRegistry() *Registry {
	return &Registry{
		RootDocument: nil,
		documents:    map[*descriptor.File]*Document{},
	}
}

// LoadGlobalConfig loads the OpenAPI document from the global config.
func (r *Registry) LoadGlobalConfig(filePath string) error {
	return r.loadFile(filePath, SourceInfo{Filename: filePath})
}

func (r *Registry) LoadFromDescriptorRegistry(reg *descriptor.Registry) error {
	return nil
}

func (r *Registry) LookupDocument(file *descriptor.File) (*Document, bool) {
	doc, ok := r.documents[file]
	return doc, ok
}

func (r *Registry) loadFile(filePath string, src SourceInfo) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	switch filepath.Ext(filePath) {
	case ".json":
		return r.loadJSON(file, src)
	case ".yaml", ".yml":
		return r.loadYAML(file, src)
	default:
		return fmt.Errorf("unrecognized/unsupported file extension: %s", filePath)
	}
}

func (r *Registry) loadYAML(reader io.Reader, src SourceInfo) error {
	var yamlContents interface{}
	if err := yaml.NewDecoder(reader).Decode(&yamlContents); err != nil {
		return fmt.Errorf("failed to decode yaml: %w", err)
	}

	jsonContents, err := json.Marshal(yamlContents)
	if err != nil {
		return fmt.Errorf("failed to JSON marshal content: %w", err)
	}

	config := &api.Config{}
	if err := protojson.Unmarshal(jsonContents, config); err != nil {
		return err
	}

	return r.processConfig(config, src)
}

func (r *Registry) loadJSON(reader io.Reader, src SourceInfo) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	config := &api.Config{}
	if err := protojson.Unmarshal(content, config); err != nil {
		return fmt.Errorf("failed to unmarshal json file: %s", err)
	}

	return r.processConfig(config, src)
}

func (r *Registry) processConfig(config *api.Config, src SourceInfo) error {
	// Ignore any file that doesn't have OpenAPI config.
	if config.Openapi == nil {
		return nil
	}

	// Process the OpenAPI document section.
	if config.Openapi.Document != nil {
		if src.IsGlobalConfigFile() { // global file
			r.RootDocument = MapDocument(config.Openapi.Document)
		} else {
			r.documents[src.ProtoFile] = MapDocument(config.Openapi.Document)
		}
	}

	// Schemas need to get validated here.

	// Service references and method references must get validated here as wel.

	return nil
}

//func (r *Registry) LookupFileDocument

func MapDocument(doc *openapi.Document) *Document {
	if doc == nil {
		return nil
	}

	result := &Document{
		ExternalDocumentation: mapExternalDoc(doc.ExternalDocs),
		Extensions:            mapExtensions(doc.Extensions),
	}

	if doc.Info != nil {
		result.Info = &Info{
			Title:          doc.Info.Title,
			Summary:        doc.Info.Summary,
			Description:    doc.Info.Description,
			TermsOfService: doc.Info.TermsOfService,
			Version:        doc.Info.Version,
		}

		if doc.Info.Contact != nil {
			result.Info.Contact = &Contact{
				Name:  doc.Info.Contact.Name,
				URL:   doc.Info.Contact.Url,
				Email: doc.Info.Contact.Email,
			}
		}

		if doc.Info.License != nil {
			result.Info.License = &License{
				Name:       doc.Info.License.Name,
				Identifier: doc.Info.License.Identifier,
				URL:        doc.Info.License.Url,
			}
		}
	}

	if doc.Servers != nil {
		for _, server := range doc.Servers {
			var vars map[string]ServerVariable
			if server.Variables != nil {
				vars = map[string]ServerVariable{}
				for name, serverVariable := range server.Variables {
					vars[name] = ServerVariable{
						Enum:        serverVariable.EnumValues,
						Default:     serverVariable.DefaultValue,
						Description: serverVariable.Description,
					}
				}
			}

			result.Servers = append(result.Servers, Server{
				URL:         server.Url,
				Description: server.Description,
				Variables:   vars,
			})
		}
	}

	if doc.Security != nil {
		result.Security = map[string][]string{}
		for _, security := range doc.Security {
			result.Security[security.Name] = security.Scopes
		}
	}

	if doc.Tags != nil {
		for _, tag := range doc.Tags {
			result.Tags = append(result.Tags, Tag{
				Name:         tag.Name,
				Description:  tag.Description,
				ExternalDocs: mapExternalDoc(tag.ExternalDocs),
			})
		}
	}

	return result
}

func mapExternalDoc(doc *openapi.ExternalDocumentation) *ExternalDocumentation {
	if doc == nil {
		return nil
	}

	return &ExternalDocumentation{
		Description: doc.Description,
		URL:         doc.Url,
	}
}

func mapExtensions(table map[string]*structpb.Value) map[string]any {
	if table == nil {
		return nil
	}

	result := make(map[string]any, len(table))
	for key, val := range table {
		result[key] = val.AsInterface()
	}

	return result
}
