package genopenapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
	RootDocument *openapiv3.Document

	documents map[*descriptor.File]*openapiv3.Document

	// messageNames holds a one to one association of message FQMN to the generated OpenAPI schema name.
	messageNames map[string]string
	// recognizedMessages holds a reference to the proto message and any matched configuration for it.
	messages map[string]struct{}
	// schemas are already processed schemas that can be readily used.
	schemas map[string]*openapiv3.Schema

	// schemas map[string]string
	// we need to populate these schemas on a need basis.
	// any time a look up is made, we find the proto definitions, load up the gateway file and build a
	// ready to use and final OpenAPI schema, this schema can later be used to inject into the OpenAPI document files.
	// another matter is that we need to already find out the name that needs to get used for this object ahead of time.
	// regardless of whether or not it gets used or not. This is to ensure two different OpenAPI files generated do not
	// use different names for the same object.
}

func NewRegistry(options *Options) *Registry {
	return &Registry{
		Options:      options,
		RootDocument: nil,
		documents:    map[*descriptor.File]*openapiv3.Document{},
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

	messages := []string{}
	err := reg.Iterate(func(filePath string, protoFile *descriptor.File) error {
		// first try to load the configFromFile file here.
		configFromFile, err := r.loadConfigForFile(filePath, protoFile)
		if err != nil {
			return fmt.Errorf("failed to load OpenAPI configs for %q: %w", filePath, err)
		}

		var doc *openapiv3.Document

		if configFromFile.OpenAPISpec != nil && configFromFile.Document != nil {
			doc, err = mapDocument(configFromFile.Document)
			if err != nil {
				return fmt.Errorf("invalid OpenAPI document in %q: %w", configFromFile.Filename, err)
			}
		}

		configFromProto, ok := proto.GetExtension(protoFile.Options, api.E_OpenapiDoc).(*openapi.Document)
		if ok && configFromProto != nil {
			docFromProto, err := mapDocument(configFromProto)
			if err != nil {
				return fmt.Errorf("invalid OpenAPI document in proto file %q: %w", filePath, err)
			}

			if doc == nil {
				doc = docFromProto
			} else if err := mergo.Merge(doc, docFromProto); err != nil {
				return fmt.Errorf("failed to merge OpenAPI config and proto documents for %q: %w", filePath, err)
			}
		}

		r.documents[protoFile] = doc

		for _, message := range protoFile.Messages {
			messages = append(messages, message.FQMN())
		}

		return nil
	})

	if err != nil {
		return err
	}

	r.messageNames = r.resolveMessageNames(messages)
	return nil
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

func (r *Registry) LookupDocument(file *descriptor.File) *openapiv3.Document {
	return r.documents[file]
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
