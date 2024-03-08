package genopenapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"dario.cat/mergo"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"

	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/configpath"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
)

// openAPIConfig is a wrapper around *api.OpenAPISpec with additional filename context.
type openAPIConfig struct {
	*api.OpenAPISpec
	Filename string
}

// Registry contains references to all configuration files.
type Registry struct {
	// Options that are shared with the generator.
	Options *Options

	// RootDocument is the global and top-level document loaded from the global config.
	RootDocument *openapiv3.ExtendedDocument

	documents map[*descriptor.File]*openapiv3.ExtendedDocument

	// schemas map[string]string
}

func NewRegistry(options *Options) *Registry {
	return &Registry{
		Options:      options,
		RootDocument: nil,
		documents:    map[*descriptor.File]*openapiv3.ExtendedDocument{},
	}
}

func (r *Registry) LoadFromDescriptorRegistry(reg *descriptor.Registry) error {
	if r.Options.GlobalOpenAPIConfigFile != "" {
		configPath := filepath.Join(r.Options.ConfigSearchPath, r.Options.GlobalOpenAPIConfigFile)
		doc, err := r.loadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to load global OpenAPI config file: %w", err)
		}

		r.RootDocument, err = mapDocument(doc.Document)
		if err != nil {
			return fmt.Errorf("invalid OpenAPI document in %q: %w", configPath, err)
		}
		// TODO: we need to process messages and services here as well.
	}

	return reg.Iterate(func(filePath string, protoFile *descriptor.File) error {
		// first try to load the configFromFile file here.
		configFromFile, err := r.loadConfigForFile(filePath, protoFile)
		if err != nil {
			return fmt.Errorf("failed to load OpenAPI configs for %q: %w", filePath, err)
		}

		var doc *openapiv3.ExtendedDocument

		if configFromFile.OpenAPISpec != nil && configFromFile.Document != nil {
			doc, err = mapDocument(configFromFile.Document)
			if err != nil {
				return fmt.Errorf("invalid OpenAPI document in %q: %w", configFromFile.Filename, err)
			}
		} else {
			doc = &openapiv3.ExtendedDocument{}
		}

		configFromProto, ok := proto.GetExtension(protoFile.Options, api.E_OpenapiDoc).(*openapi.Document)
		if ok && configFromProto != nil {
			docFromProto, err := mapDocument(configFromProto)
			if err != nil {
				return fmt.Errorf("invalid OpenAPI document in proto file %q: %w", filePath, err)
			}
			if err := mergo.Merge(doc, docFromProto); err != nil {
				return fmt.Errorf("failed to merge OpenAPI config and proto documents for %q: %w", filePath, err)
			}
		}

		r.documents[protoFile] = doc
		return nil
	})
}

func (r *Registry) loadConfigForFile(protoFilePath string, file *descriptor.File) (openAPIConfig, error) {
	// TODO: allow the plugin to set the config path
	result := openAPIConfig{}
	if r.Options.OpenAPIConfigFilePattern == "" {
		return result, nil
	}

	configPath, err := configpath.Build(protoFilePath, r.Options.OpenAPIConfigFilePattern)
	if err != nil {
		return result, fmt.Errorf("failed to determine config file path: %w", err)
	}

	for _, ext := range [...]string{"yaml", "yml", "json"} {
		configFilePath := filepath.Join(r.Options.ConfigSearchPath, configPath+"."+ext)

		if _, err := os.Stat(configFilePath); err != nil {
			if os.IsNotExist(err) {
				grpclog.Infof("looked for file %s, it was not found", configFilePath)
				continue
			}

			return result, fmt.Errorf("failed to stat file '%s': %w", configFilePath, err)
		}

		// file exists, try to load it.
		config, err := r.loadFile(configFilePath)
		if err != nil {
			return result, fmt.Errorf("failed to load %s: %w", configFilePath, err)
		}

		return openAPIConfig{OpenAPISpec: config, Filename: configFilePath}, nil
	}

	return result, nil
}

func (r *Registry) LookupDocument(file *descriptor.File) (*openapiv3.ExtendedDocument, bool) {
	doc, ok := r.documents[file]
	return doc, ok
}

func (r *Registry) loadFile(filePath string) (*api.OpenAPISpec, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	switch filepath.Ext(filePath) {
	case ".json":
		return r.loadJSON(file)
	case ".yaml", ".yml":
		return r.loadYAML(file)
	default:
		return nil, fmt.Errorf("unrecognized/unsupported file extension: %s", filePath)
	}
}

func (r *Registry) loadYAML(reader io.Reader) (*api.OpenAPISpec, error) {
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

func (r *Registry) loadJSON(reader io.Reader) (*api.OpenAPISpec, error) {
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

func mapDocument(doc *openapi.Document) (*openapiv3.ExtendedDocument, error) {
	if doc == nil {
		return nil, nil
	}

	extensions, err := mapExtensions(doc.Extensions)
	if err != nil {
		return nil, err
	}

	info, err := mapInfo(doc.Info)
	if err != nil {
		return nil, fmt.Errorf("invalid info object: %w", err)
	}

	result := &openapiv3.ExtendedDocument{
		Object: openapiv3.Document{
			Info: info,
		},
		Extensions: extensions,
	}

	result.Object.ExternalDocumentation, err = mapExternalDoc(doc.ExternalDocs)
	if err != nil {
		return nil, fmt.Errorf("invalid external doc: %w", err)
	}

	result.Object.Servers, err = mapServers(doc.Servers)
	if err != nil {
		return nil, fmt.Errorf("invalid servers list: %w", err)
	}

	if doc.Security != nil {
		result.Object.Security = map[string][]string{}
		for _, security := range doc.Security {
			result.Object.Security[security.Name] = security.Scopes
		}
	}

	result.Object.Tags, err = mapTags(doc.Tags)
	if err != nil {
		return nil, fmt.Errorf("invalid tags list: %w", err)
	}

	return result, nil
}

func mapExternalDoc(doc *openapi.ExternalDocumentation) (*openapiv3.ExtendedExternalDocumentation, error) {
	if doc == nil {
		return nil, nil
	}

	extensions, err := mapExtensions(doc.Extensions)
	if err != nil {
		return nil, err
	}

	return &openapiv3.ExtendedExternalDocumentation{
		Object: openapiv3.ExternalDocumentation{
			Description: doc.Description,
			URL:         doc.Url,
		},
		Extensions: extensions,
	}, nil
}

func mapExtensions(table map[string]*structpb.Value) (map[string]any, error) {
	if table == nil {
		return nil, nil
	}

	result := make(map[string]any, len(table))
	for key, val := range table {
		if !strings.HasPrefix(key, "x-") {
			return nil, fmt.Errorf("extension keys must begin with prefix 'x-', value %q is not accepted", key)
		}
		result[key] = val.AsInterface()
	}

	return result, nil
}

func mapInfo(info *openapi.Info) (*openapiv3.ExtendedInfo, error) {
	if info == nil {
		return nil, nil
	}

	extensions, err := mapExtensions(info.Extensions)
	if err != nil {
		return nil, err
	}

	result := &openapiv3.ExtendedInfo{
		Object: openapiv3.Info{
			Title:          info.Title,
			Summary:        info.Summary,
			Description:    info.Description,
			TermsOfService: info.TermsOfService,
			Version:        info.Version,
		},
		Extensions: extensions,
	}

	if info.Contact != nil {
		extensions, err = mapExtensions(info.Contact.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid contact object: %w", err)
		}

		result.Object.Contact = &openapiv3.ExtendedContact{
			Object: openapiv3.Contact{
				Name:  info.Contact.Name,
				URL:   info.Contact.Url,
				Email: info.Contact.Email,
			},
			Extensions: extensions,
		}
	}

	if info.License != nil {
		extensions, err = mapExtensions(info.License.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid contact object: %w", err)
		}

		result.Object.License = &openapiv3.ExtendedLicense{
			Object: openapiv3.License{
				Name:       info.License.Name,
				Identifier: info.License.Identifier,
				URL:        info.License.Url,
			},
			Extensions: extensions,
		}
	}

	return result, nil
}

func mapTags(tags []*openapi.Tag) ([]openapiv3.ExtendedTag, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	// defining these variables outside of the for loop to reuse them.
	var extensions openapiv3.Extensions
	var externalDocs *openapiv3.ExtendedExternalDocumentation
	var err error

	result := make([]openapiv3.ExtendedTag, len(tags))
	for index, tag := range tags {
		extensions, err = mapExtensions(tag.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid tag at index %d: %w", index, err)
		}

		externalDocs, err = mapExternalDoc(tag.ExternalDocs)
		if err != nil {
			return nil, fmt.Errorf("invalid external doc in tag at index %d: %w", index, err)
		}

		result[index] = openapiv3.ExtendedTag{
			Object: openapiv3.Tag{
				Name:         tag.Name,
				Description:  tag.Description,
				ExternalDocs: externalDocs,
			},
			Extensions: extensions,
		}
	}

	return result, nil
}

func mapServers(servers []*openapi.Server) ([]openapiv3.ExtendedServer, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	// defining these variables outside of the for loop to reuse them.
	var extensions, serverVarExtensions openapiv3.Extensions
	var err error

	result := make([]openapiv3.ExtendedServer, len(servers))
	for index, server := range servers {
		extensions, err = mapExtensions(server.Extensions)
		if err != nil {
			return nil, fmt.Errorf("invalid server at index %d: %w", index, err)
		}

		var vars map[string]openapiv3.ExtendedServerVariable
		if server.Variables != nil {
			vars = map[string]openapiv3.ExtendedServerVariable{}
			for name, serverVariable := range server.Variables {
				serverVarExtensions, err = mapExtensions(serverVariable.Extensions)
				if err != nil {
					return nil, fmt.Errorf("invalid server variable %q: %w", name, err)
				}

				vars[name] = openapiv3.ExtendedServerVariable{
					Object: openapiv3.ServerVariable{
						Enum:        serverVariable.EnumValues,
						Default:     serverVariable.DefaultValue,
						Description: serverVariable.Description,
					},
					Extensions: serverVarExtensions,
				}
			}
		}

		result[index] = openapiv3.ExtendedServer{
			Object: openapiv3.Server{
				URL:         server.Url,
				Description: server.Description,
				Variables:   vars,
			},
			Extensions: extensions,
		}
	}

	return result, nil
}
