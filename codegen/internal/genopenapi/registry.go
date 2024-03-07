package genopenapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/configpath"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

type SourceInfo struct {
	Filename  string
	ProtoFile *descriptor.File
}

func (s *SourceInfo) IsGlobalConfigFile() bool {
	return s.ProtoFile == nil
}

// Registry contains references to all configuration files.
type Registry struct {
	// Options that are shared with the generator.
	Options *Options

	// RootDocument is the global and top-level document loaded from the global config.
	RootDocument *openapiv3.Document

	documents map[*descriptor.File]*openapiv3.Document

	// schemas map[string]string
}

func NewRegistry(options *Options) *Registry {
	return &Registry{
		Options:      options,
		RootDocument: nil,
		documents:    map[*descriptor.File]*openapiv3.Document{},
	}
}

// LoadGlobalConfig loads the OpenAPI document from the global config.
//func (r *Registry) LoadGlobalConfig(filePath string) error {
//}

func (r *Registry) LoadFromDescriptorRegistry(reg *descriptor.Registry) error {
	if r.Options.GlobalOpenAPIConfigFile != "" {
		configPath := filepath.Join(r.Options.ConfigSearchPath, r.Options.GlobalOpenAPIConfigFile)
		doc, err := r.loadFile(configPath, SourceInfo{Filename: configPath})
		if err != nil {
			return fmt.Errorf("failed to load global OpenAPI config file: %w", err)
		}

		r.RootDocument = MapDocument(doc.Document)
		// TODO: we need to process messages and services here as well.
	}

	return reg.Iterate(func(filePath string, protoFile *descriptor.File) error {
		// first try to load the config file here.
		configSpec, err := r.loadConfigForFile(filePath, protoFile)
		if err != nil {
			return fmt.Errorf("failed to load OpenAPI configs for %q: %w", filePath, err)
		}
		var doc *openapiv3.Document

		if configSpec != nil && configSpec.Document != nil {
			doc = MapDocument(configSpec.Document)
		} else {
			doc = &openapiv3.Document{}
		}

		protoSpec, ok := proto.GetExtension(protoFile.Options, api.E_OpenapiDoc).(*openapi.Document)
		if ok && protoSpec != nil {
			if err := mergo.Merge(doc, MapDocument(protoSpec)); err != nil {
				return fmt.Errorf("failed to merge OpenAPI config and proto documents for %q: %w", filePath, err)
			}
		}

		r.documents[protoFile] = doc
		return nil
	})
}

func (r *Registry) loadConfigForFile(protoFilePath string, file *descriptor.File) (*api.OpenAPISpec, error) {
	// TODO: allow the plugin to set the config path
	if r.Options.OpenAPIConfigFilePattern == "" {
		return nil, nil
	}

	configPath, err := configpath.Build(protoFilePath, r.Options.OpenAPIConfigFilePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to determine config file path: %w", err)
	}

	for _, ext := range [...]string{"yaml", "yml", "json"} {
		configFilePath := filepath.Join(r.Options.ConfigSearchPath, configPath+"."+ext)

		if _, err := os.Stat(configFilePath); err != nil {
			if os.IsNotExist(err) {
				grpclog.Infof("looked for file %s, it was not found", configFilePath)
				continue
			}

			return nil, fmt.Errorf("failed to stat file '%s': %w", configFilePath, err)
		}

		// file exists, try to load it.
		config, err := r.loadFile(configFilePath, SourceInfo{
			Filename:  configFilePath,
			ProtoFile: file,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", configFilePath, err)
		}

		return config, nil
	}

	return nil, nil
}

func (r *Registry) LookupDocument(file *descriptor.File) (*openapiv3.Document, bool) {
	doc, ok := r.documents[file]
	return doc, ok
}

func (r *Registry) loadFile(filePath string, src SourceInfo) (*api.OpenAPISpec, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	switch filepath.Ext(filePath) {
	case ".json":
		return r.loadJSON(file, src)
	case ".yaml", ".yml":
		return r.loadYAML(file, src)
	default:
		return nil, fmt.Errorf("unrecognized/unsupported file extension: %s", filePath)
	}
}

func (r *Registry) loadYAML(reader io.Reader, src SourceInfo) (*api.OpenAPISpec, error) {
	var yamlContents interface{}
	if err := yaml.NewDecoder(reader).Decode(&yamlContents); err != nil {
		return nil, fmt.Errorf("failed to decode yaml: %w", err)
	}

	jsonContents, err := json.Marshal(yamlContents)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON marshal content: %w", err)
	}

	config := &api.Config{}
	if err := protojson.Unmarshal(jsonContents, config); err != nil {
		return nil, err
	}

	return config.Openapi, nil
}

func (r *Registry) loadJSON(reader io.Reader, src SourceInfo) (*api.OpenAPISpec, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}

	config := &api.Config{}
	if err := protojson.Unmarshal(content, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json file: %s", err)
	}

	return config.Openapi, nil
}

//func (r *Registry) processConfig(config *api.Config, src SourceInfo) error {
//  // Ignore any file that doesn't have OpenAPI config.
//  if config.Openapi == nil {
//    return nil
//  }

//  // Process the OpenAPI document section.
//  if config.Openapi.Document != nil {
//    if src.IsGlobalConfigFile() { // global file
//      r.RootDocument = MapDocument(config.Openapi.Document)
//    } else {
//      r.documents[src.ProtoFile] = MapDocument(config.Openapi.Document)
//    }
//  }

//  // Schemas need to get validated here.

//  // Service references and method references must get validated here as wel.

//  return nil
//}

func MapDocument(doc *openapi.Document) *openapiv3.Document {
	if doc == nil {
		return nil
	}

	result := &openapiv3.Document{
		ExternalDocumentation: mapExternalDoc(doc.ExternalDocs),
		Extensions:            mapExtensions(doc.Extensions),
	}

	if doc.Info != nil {
		result.Info = &openapiv3.Info{
			Title:          doc.Info.Title,
			Summary:        doc.Info.Summary,
			Description:    doc.Info.Description,
			TermsOfService: doc.Info.TermsOfService,
			Version:        doc.Info.Version,
		}

		if doc.Info.Contact != nil {
			result.Info.Contact = &openapiv3.Contact{
				Name:  doc.Info.Contact.Name,
				URL:   doc.Info.Contact.Url,
				Email: doc.Info.Contact.Email,
			}
		}

		if doc.Info.License != nil {
			result.Info.License = &openapiv3.License{
				Name:       doc.Info.License.Name,
				Identifier: doc.Info.License.Identifier,
				URL:        doc.Info.License.Url,
			}
		}
	}

	if doc.Servers != nil {
		for _, server := range doc.Servers {
			var vars map[string]openapiv3.ServerVariable
			if server.Variables != nil {
				vars = map[string]openapiv3.ServerVariable{}
				for name, serverVariable := range server.Variables {
					vars[name] = openapiv3.ServerVariable{
						Enum:        serverVariable.EnumValues,
						Default:     serverVariable.DefaultValue,
						Description: serverVariable.Description,
					}
				}
			}

			result.Servers = append(result.Servers, openapiv3.Server{
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
			result.Tags = append(result.Tags, openapiv3.Tag{
				Name:         tag.Name,
				Description:  tag.Description,
				ExternalDocs: mapExternalDoc(tag.ExternalDocs),
			})
		}
	}

	return result
}

func mapExternalDoc(doc *openapi.ExternalDocumentation) *openapiv3.ExternalDocumentation {
	if doc == nil {
		return nil
	}

	return &openapiv3.ExternalDocumentation{
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
		result["x-"+key] = val.AsInterface()
	}

	return result
}
