package genopenapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"

	"github.com/meshapi/grpc-rest-gateway/api"
	"github.com/meshapi/grpc-rest-gateway/api/openapi"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/configpath"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/internal"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genopenapi/openapimap"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/openapiv3"
)

func (g *Generator) readOpenAPISeedFile() (map[string]any, error) {
	seedFile, err := os.Open(g.configFilePath(g.Options.OpenAPISeedFile))
	if err != nil {
		return nil, fmt.Errorf("failed to open seed file: %w", err)
	}
	defer seedFile.Close()

	var value map[string]any

	extension := filepath.Ext(g.Options.OpenAPISeedFile)
	switch strings.ToLower(extension) {
	case ".json":
		err = json.NewDecoder(seedFile).Decode(&value)
	case ".yml", ".yaml":
		err = yaml.NewDecoder(seedFile).Decode(&value)
	default:
		return nil, fmt.Errorf("could not recognize file type of the OpenAPI seed file %q", g.OpenAPISeedFile)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI seed file %q: %w", g.Options.OpenAPISeedFile, err)
	}

	return value, nil
}

// configFilePath uses the config search path to find the final path for a config file.
func (g *Generator) configFilePath(fileName string) string {
	if filepath.IsAbs(fileName) {
		return fileName
	}
	return filepath.Join(g.ConfigSearchPath, fileName)
}

func (g *Generator) loadFromDescriptorRegistry() error {
	if g.GlobalOpenAPIConfigFile != "" {
		configPath := g.configFilePath(g.GlobalOpenAPIConfigFile)
		doc, err := g.loadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to load global OpenAPI config file: %w", err)
		}

		g.rootDocument.Document, err = openapimap.Document(doc.Document)
		if err != nil {
			return fmt.Errorf("invalid OpenAPI document in %q: %w", configPath, err)
		}

		g.rootDocument.DefaultResponses, err = internal.MapDefaultResponses(
			doc.GetDocument().GetConfig().GetDefaultResponses())
		if err != nil {
			return fmt.Errorf("failed to map default responses to OpenAPI response in %q: %w", configPath, err)
		}

		err = g.addMessageConfigs(doc.Messages, internal.SourceInfo{Filename: configPath})
		if err != nil {
			return fmt.Errorf("failed to process message configs defined in %q: %w", configPath, err)
		}

		err = g.addEnumConfigs(doc.Enums, internal.SourceInfo{Filename: configPath})
		if err != nil {
			return fmt.Errorf("failed to process enum configs defined in %q: %w", configPath, err)
		}

		err = g.addServiceConfigs(doc.Services, internal.SourceInfo{Filename: configPath})
		if err != nil {
			return fmt.Errorf("failed to process service configs defined in %q: %w", configPath, err)
		}
	}

	protoTypes := []internal.ProtoTypeName{}
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

			source := internal.SourceInfo{
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

		var configFromProto *openapi.Document
		var ok bool
		if protoFile.Options != nil && proto.HasExtension(protoFile.Options, api.E_OpenapiDoc) {
			configFromProto, ok = proto.GetExtension(protoFile.Options, api.E_OpenapiDoc).(*openapi.Document)
		}
		if ok && configFromProto != nil {
			docFromProto, err := openapimap.Document(configFromProto)
			if err != nil {
				return fmt.Errorf("invalid OpenAPI document in proto file %q: %w", filePath, err)
			}

			if doc == nil {
				doc = docFromProto
			} else if err := g.mergeObjects(doc, docFromProto); err != nil {
				return fmt.Errorf("failed to merge OpenAPI config and proto documents for %q: %w", filePath, err)
			}
		}

		defaultResponsesSpec := internal.MergeDefaultResponseSpec(
			configFromProto.GetConfig().GetDefaultResponses(),
			configFromFile.GetDocument().GetConfig().GetDefaultResponses())
		defaultResponses, err := internal.MapDefaultResponses(defaultResponsesSpec)
		if err != nil {
			return fmt.Errorf("failed to map config response map to OpenAPI response map: %w", err)
		}
		defaultResponses = internal.MergeDefaultResponses(defaultResponses, g.rootDocument.DefaultResponses)
		g.files[protoFile] = internal.OpenAPIDocument{
			Document:         doc,
			DefaultResponses: defaultResponses,
		}

		for _, message := range protoFile.Messages {
			protoTypes = append(protoTypes, internal.ProtoTypeName{FQN: message.FQMN(), OuterLength: uint8(len(message.Outers))})
		}

		for _, enum := range protoFile.Enums {
			protoTypes = append(protoTypes, internal.ProtoTypeName{FQN: enum.FQEN(), OuterLength: uint8(len(enum.Outers))})
		}

		return nil
	})

	if err != nil {
		return err
	}

	g.schemaNames = g.resolveTypeNames(protoTypes)
	return nil
}

func (g *Generator) addMessageConfigs(configs []*api.OpenAPIMessageSpec, src internal.SourceInfo) error {
	for _, messageConfig := range configs {
		// Resolve the selector to an absolute path.
		if strings.HasPrefix(messageConfig.Selector, "~.") {
			if src.ProtoPackage == "" {
				return fmt.Errorf("no proto package context is available, cannot use relative selector: %s", messageConfig.Selector)
			}

			messageConfig.Selector = "." + src.ProtoPackage + messageConfig.Selector[1:]
		}

		// assert that selector resolves to a proto message.
		msg, err := g.registry.LookupMessage(src.ProtoPackage, messageConfig.Selector)
		if err != nil {
			return fmt.Errorf(
				"could not find proto message %q referenced in file: %s", messageConfig.Selector, src.Filename)
		}

		if existingConfig, alreadyExists := g.messages[msg.FQMN()]; alreadyExists {
			return fmt.Errorf(
				"multiple external OpenAPI configurations for message %q: both %q and %q contain bindings for this selector",
				messageConfig.Selector, existingConfig.Filename, src.Filename)
		}

		g.messages[msg.FQMN()] = &internal.OpenAPIMessageSpec{
			OpenAPIMessageSpec: messageConfig,
			SourceInfo:         src,
		}
	}

	return nil
}

func (g *Generator) addEnumConfigs(configs []*api.OpenAPIEnumSpec, src internal.SourceInfo) error {
	for _, enumConfig := range configs {
		// Resolve the selector to an absolute path.
		if strings.HasPrefix(enumConfig.Selector, "~.") {
			if src.ProtoPackage == "" {
				return fmt.Errorf("no proto package context is available, cannot use relative selector: %s", enumConfig.Selector)
			}
			enumConfig.Selector = "." + src.ProtoPackage + enumConfig.Selector[1:]
		}

		// assert that selector resolves to a proto enum.
		enum, err := g.registry.LookupEnum(src.ProtoPackage, enumConfig.Selector)
		if err != nil {
			return fmt.Errorf(
				"could not find proto enum %q referenced in file: %s", enumConfig.Selector, src.Filename)
		}

		if existingConfig, alreadyExists := g.enums[enum.FQEN()]; alreadyExists {
			return fmt.Errorf(
				"multiple external OpenAPI configurations for enum %q: both %q and %q contain bindings for this selector",
				enumConfig.Selector, existingConfig.Filename, src.Filename)
		}

		g.enums[enum.FQEN()] = &internal.OpenAPIEnumSpec{
			OpenAPIEnumSpec: enumConfig,
			SourceInfo:      src,
		}
	}

	return nil
}

func (g *Generator) addServiceConfigs(configs []*api.OpenAPIServiceSpec, src internal.SourceInfo) error {
	for _, serviceConfig := range configs {
		// Resolve the selector to an absolute path.
		if strings.HasPrefix(serviceConfig.Selector, "~.") {
			if src.ProtoPackage == "" {
				return fmt.Errorf("no proto package context is available, cannot use relative selector: %s", serviceConfig.Selector)
			}
			serviceConfig.Selector = src.ProtoPackage + serviceConfig.Selector[1:]
		}

		if !strings.HasPrefix(serviceConfig.Selector, ".") {
			serviceConfig.Selector = "." + serviceConfig.Selector
		}

		if existingConfig, alreadyExists := g.services[serviceConfig.Selector]; alreadyExists {
			return fmt.Errorf(
				"multiple external OpenAPI configurations for service %q: both %q and %q contain bindings for this selector",
				serviceConfig.Selector, existingConfig.Filename, src.Filename)
		}

		g.services[serviceConfig.Selector] = &internal.OpenAPIServiceSpec{
			OpenAPIServiceSpec: serviceConfig,
			SourceInfo:         src,
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

	g.schemas[fqen] = internal.OpenAPISchema{Schema: result}
	return result, nil
}

func (g *Generator) getSchemaForMessage(protoPackage, fqmn string) (internal.OpenAPISchema, error) {
	// first look up the cache
	if result, alreadyProcessed := g.schemas[fqmn]; alreadyProcessed {
		return result, nil
	}

	// pull up the proto message options and configs, render the schema and then merge if needed.
	message, err := g.registry.LookupMessage(protoPackage, fqmn)
	if err != nil {
		return internal.OpenAPISchema{}, fmt.Errorf("failed to find proto message: %w", err)
	}

	result, err := g.renderMessageSchema(message)
	if err != nil {
		return internal.OpenAPISchema{}, fmt.Errorf("failed to render message: %w", err)
	}

	g.schemas[fqmn] = result
	return result, nil
}

func (g *Generator) tagsForService(service *descriptor.Service) ([]*openapiv3.Tag, error) {
	opts, ok := g.services[service.FQSN()]
	if ok {
		tags, err := openapimap.Tags(opts.Document.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to map tags: %w", err)
		}

		return tags, nil
	}

	if service.Options == nil || !proto.HasExtension(service.Options, api.E_OpenapiServiceDoc) {
		return nil, nil
	}
	protoOptions, ok := proto.GetExtension(service.Options, api.E_OpenapiServiceDoc).(*openapi.Document)
	if ok && protoOptions != nil {
		tags, err := openapimap.Tags(protoOptions.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to map tags: %w", err)
		}
		return tags, nil
	}

	return nil, nil
}

func (g *Generator) loadConfigForFile(protoFilePath string, file *descriptor.File) (internal.OpenAPISpec, error) {
	// TODO: allow the plugin to set the config path
	result := internal.OpenAPISpec{}
	if g.OpenAPIConfigFilePattern == "" {
		return result, nil
	}

	configPath, err := configpath.Build(protoFilePath, g.OpenAPIConfigFilePattern)
	if err != nil {
		return result, fmt.Errorf("failed to determine config file path: %w", err)
	}

	for _, ext := range [...]string{"yaml", "yml", "json"} {
		configFilePath := g.configFilePath(configPath + "." + ext)

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

		return internal.OpenAPISpec{OpenAPISpec: config, Filename: configFilePath}, nil
	}

	return result, nil
}

func (g *Generator) LookupFile(file *descriptor.File) internal.OpenAPIDocument {
	return g.files[file]
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
