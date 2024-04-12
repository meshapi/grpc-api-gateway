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
	"gopkg.in/yaml.v3"

	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/configpath"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/openapimap"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
)

type dependencyKind uint8

const (
	dependencyKindEnum dependencyKind = iota
	dependencyKindMessage
)

// openAPIConfig is a wrapper around *api.OpenAPISpec with additional filename context.
type openAPIConfig struct {
	*api.OpenAPISpec
	Filename string
}

type sourceInfo struct {
	Filename     string
	ProtoPackage string
}

type openAPIMessageConfig struct {
	*api.OpenAPIMessageSpec
	sourceInfo
}

type openAPIServiceConfig struct {
	*api.OpenAPIServiceSpec
	sourceInfo
}

type openAPIEnumConfig struct {
	*api.OpenAPIEnumSpec
	sourceInfo
}

type schemaDependency struct {
	FQN  string
	Kind dependencyKind
}

func (s schemaDependency) IsSet() bool {
	return s.FQN != ""
}

type openAPISchemaConfig struct {
	// Schema is the already mapped and processed OpenAPI schema for a proto message/enum.
	Schema *openapiv3.Schema
	// Dependencies is the list of enum or proto message dependencies that must be included in the same
	// OpenAPI document.
	Dependencies []schemaDependency
}

func (g *Generator) LoadFromDescriptorRegistry() error {
	if g.GlobalOpenAPIConfigFile != "" {
		configPath := filepath.Join(g.ConfigSearchPath, g.GlobalOpenAPIConfigFile)
		doc, err := g.loadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to load global OpenAPI config file: %w", err)
		}

		g.rootDocument, err = openapimap.Document(doc.Document)
		if err != nil {
			return fmt.Errorf("invalid OpenAPI document in %q: %w", configPath, err)
		}

		err = g.addMessageConfigs(doc.Messages, sourceInfo{Filename: configPath})
		if err != nil {
			return fmt.Errorf("failed to process message configs defined in %q: %w", configPath, err)
		}

		err = g.addEnumConfigs(doc.Enums, sourceInfo{Filename: configPath})
		if err != nil {
			return fmt.Errorf("failed to process enum configs defined in %q: %w", configPath, err)
		}

		err = g.addServiceConfigs(doc.Services, sourceInfo{Filename: configPath})
		if err != nil {
			return fmt.Errorf("failed to process service configs defined in %q: %w", configPath, err)
		}
	}

	messages := []string{}
	err := g.registry.Iterate(func(filePath string, protoFile *descriptor.File) error {
		// first try to load the configFromFile file here.
		configFromFile, err := g.loadConfigForFile(filePath, protoFile)
		if err != nil {
			return fmt.Errorf("failed to load OpenAPI configs for %q: %w", filePath, err)
		}

		var doc *openapiv3.Document

		if configFromFile.OpenAPISpec != nil && configFromFile.Document != nil {
			doc, err = openapimap.Document(configFromFile.Document)
			if err != nil {
				return fmt.Errorf("invalid OpenAPI document in %q: %w", configFromFile.Filename, err)
			}

			source := sourceInfo{
				Filename:     configFromFile.Filename,
				ProtoPackage: protoFile.GetPackage(),
			}

			if err := g.addMessageConfigs(configFromFile.Messages, source); err != nil {
				return fmt.Errorf("failed to process message configs defined in %q: %w", configFromFile.Filename, err)
			}
			if err := g.addEnumConfigs(configFromFile.Enums, source); err != nil {
				return fmt.Errorf("failed to process enum configs defined in %q: %w", configFromFile.Filename, err)
			}
			if err := g.addServiceConfigs(configFromFile.Services, source); err != nil {
				return fmt.Errorf("failed to process service configs defined in %q: %w", configFromFile.Filename, err)
			}
		}

		configFromProto, ok := proto.GetExtension(protoFile.Options, api.E_OpenapiDoc).(*openapi.Document)
		if ok && configFromProto != nil {
			docFromProto, err := openapimap.Document(configFromProto)
			if err != nil {
				return fmt.Errorf("invalid OpenAPI document in proto file %q: %w", filePath, err)
			}

			if doc == nil {
				doc = docFromProto
			} else if err := mergo.Merge(doc, docFromProto); err != nil {
				return fmt.Errorf("failed to merge OpenAPI config and proto documents for %q: %w", filePath, err)
			}
		}

		g.documents[protoFile] = doc

		for _, message := range protoFile.Messages {
			messages = append(messages, message.FQMN())
		}

		for _, enum := range protoFile.Enums {
			messages = append(messages, enum.FQEN())
		}

		return nil
	})

	if err != nil {
		return err
	}

	g.schemaNames = g.resolveMessageNames(messages)
	return nil
}

func (g *Generator) addMessageConfigs(configs []*api.OpenAPIMessageSpec, src sourceInfo) error {
	for _, messageConfig := range configs {
		// Resolve the selector to an absolute path.
		if strings.HasPrefix(messageConfig.Selector, ".") {
			if src.ProtoPackage == "" {
				return fmt.Errorf("no proto package context is available, cannot use relative selector: %s", messageConfig.Selector)
			}
			messageConfig.Selector = src.ProtoPackage + messageConfig.Selector
		}

		messageConfig.Selector = "." + messageConfig.Selector

		// assert that selector resolves to a proto message.
		if _, err := g.registry.LookupMessage("", messageConfig.Selector); err != nil {
			return fmt.Errorf(
				"could not find proto message %q referenced in file: %s", messageConfig.Selector, src.Filename)
		}

		if existingConfig, alreadyExists := g.messages[messageConfig.Selector]; alreadyExists {
			return fmt.Errorf(
				"multiple external OpenAPI configurations for message %q: both %q and %q contain bindings for this selector",
				messageConfig.Selector, existingConfig.Filename, src.Filename)
		}

		g.messages[messageConfig.Selector] = &openAPIMessageConfig{
			OpenAPIMessageSpec: messageConfig,
			sourceInfo:         src,
		}
	}

	return nil
}

func (g *Generator) addEnumConfigs(configs []*api.OpenAPIEnumSpec, src sourceInfo) error {
	for _, enumConfig := range configs {
		// Resolve the selector to an absolute path.
		if strings.HasPrefix(enumConfig.Selector, ".") {
			if src.ProtoPackage == "" {
				return fmt.Errorf("no proto package context is available, cannot use relative selector: %s", enumConfig.Selector)
			}
			enumConfig.Selector = src.ProtoPackage + enumConfig.Selector
		}

		enumConfig.Selector = "." + enumConfig.Selector

		// assert that selector resolves to a proto enum.
		if _, err := g.registry.LookupEnum("", enumConfig.Selector); err != nil {
			return fmt.Errorf(
				"could not find proto enum %q referenced in file: %s", enumConfig.Selector, src.Filename)
		}

		if existingConfig, alreadyExists := g.enums[enumConfig.Selector]; alreadyExists {
			return fmt.Errorf(
				"multiple external OpenAPI configurations for enum %q: both %q and %q contain bindings for this selector",
				enumConfig.Selector, existingConfig.Filename, src.Filename)
		}

		g.enums[enumConfig.Selector] = &openAPIEnumConfig{
			OpenAPIEnumSpec: enumConfig,
			sourceInfo:      src,
		}
	}

	return nil
}

func (g *Generator) addServiceConfigs(configs []*api.OpenAPIServiceSpec, src sourceInfo) error {
	for _, serviceConfig := range configs {
		// Resolve the selector to an absolute path.
		if strings.HasPrefix(serviceConfig.Selector, ".") {
			if src.ProtoPackage == "" {
				return fmt.Errorf("no proto package context is available, cannot use relative selector: %s", serviceConfig.Selector)
			}
			serviceConfig.Selector = src.ProtoPackage + serviceConfig.Selector
		}

		serviceConfig.Selector = "." + serviceConfig.Selector

		if existingConfig, alreadyExists := g.services[serviceConfig.Selector]; alreadyExists {
			return fmt.Errorf(
				"multiple external OpenAPI configurations for message %q: both %q and %q contain bindings for this selector",
				serviceConfig.Selector, existingConfig.Filename, src.Filename)
		}

		g.services[serviceConfig.Selector] = &openAPIServiceConfig{
			OpenAPIServiceSpec: serviceConfig,
			sourceInfo:         src,
		}
	}

	return nil
}

func (g *Generator) getSchemaForEnum(protoPackage, fqen string) (*openapiv3.Schema, error) {
	// first look up the cache
	if result, alreadyProcessed := g.schemas[fqen]; alreadyProcessed {
		return result.Schema, nil
	}

	enum, err := g.registry.LookupEnum(protoPackage, fqen)
	if err != nil {
		return nil, fmt.Errorf("failed to find enum: %w", err)
	}

	result, err := g.renderEnumSchema(enum)
	if err != nil {
		return nil, fmt.Errorf("failed to render enum: %w", err)
	}

	g.schemas[fqen] = openAPISchemaConfig{Schema: result}
	return result, nil
}

// lookUpFieldConfig looks up field configuration by searching the configs and then the proto extension.
//
// TODO: revisit this method for performance.
func (g *Generator) lookUpFieldConfig(field *descriptor.Field) *openapi.FieldConfiguration {
	if config, ok := g.messages[field.Message.FQMN()]; ok {
		if fieldConfig := config.Fields[field.GetName()]; fieldConfig != nil && fieldConfig.Config != nil {
			return fieldConfig.Config
		}
	}

	protoConfig, ok := proto.GetExtension(field.Options, api.E_OpenapiField).(*openapi.Schema)
	if ok && protoConfig != nil && protoConfig.Config != nil {
		return protoConfig.Config
	}

	return nil
}

func (g *Generator) getSchemaForMessage(protoPackage, fqmn string) (openAPISchemaConfig, error) {
	// first look up the cache
	if result, alreadyProcessed := g.schemas[fqmn]; alreadyProcessed {
		return result, nil
	}

	// pull up the proto message options and configs, render the schema and then merge if needed.
	message, err := g.registry.LookupMessage(protoPackage, fqmn)
	if err != nil {
		return openAPISchemaConfig{}, fmt.Errorf("failed to find proto message: %w", err)
	}

	result, err := g.renderMessageSchema(message)
	if err != nil {
		return openAPISchemaConfig{}, fmt.Errorf("failed to render message: %w", err)
	}

	g.schemas[fqmn] = result
	return result, nil
}

func (g *Generator) loadConfigForFile(protoFilePath string, file *descriptor.File) (openAPIConfig, error) {
	// TODO: allow the plugin to set the config path
	result := openAPIConfig{}
	if g.OpenAPIConfigFilePattern == "" {
		return result, nil
	}

	configPath, err := configpath.Build(protoFilePath, g.OpenAPIConfigFilePattern)
	if err != nil {
		return result, fmt.Errorf("failed to determine config file path: %w", err)
	}

	for _, ext := range [...]string{"yaml", "yml", "json"} {
		configFilePath := filepath.Join(g.ConfigSearchPath, configPath+"."+ext)

		if _, err := os.Stat(configFilePath); err != nil {
			if os.IsNotExist(err) {
				grpclog.Infof("looked for file %s, it was not found", configFilePath)
				continue
			}

			return result, fmt.Errorf("failed to stat file '%s': %w", configFilePath, err)
		}

		// file exists, try to load it.
		config, err := g.loadFile(configFilePath)
		if err != nil {
			return result, fmt.Errorf("failed to load %s: %w", configFilePath, err)
		}

		return openAPIConfig{OpenAPISpec: config, Filename: configFilePath}, nil
	}

	return result, nil
}

func (g *Generator) LookupDocument(file *descriptor.File) *openapiv3.Document {
	return g.documents[file]
}

func (g *Generator) loadFile(filePath string) (*api.OpenAPISpec, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	switch filepath.Ext(filePath) {
	case ".json":
		return g.loadJSON(file)
	case ".yaml", ".yml":
		return g.loadYAML(file)
	default:
		return nil, fmt.Errorf("unrecognized/unsupported file extension: %s", filePath)
	}
}

func (g *Generator) loadYAML(reader io.Reader) (*api.OpenAPISpec, error) {
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

func (g *Generator) loadJSON(reader io.Reader) (*api.OpenAPISpec, error) {
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
