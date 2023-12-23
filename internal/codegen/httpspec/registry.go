package httpspec

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

var (
	selectorPattern = regexp.MustCompile(`^\w+(?:[.]\w+)+$`)
)

type SourceInfo struct {
	Filename     string
	ProtoPackage string
}

type EndpointSpec struct {
	Binding    *EndpointBinding
	SourceInfo SourceInfo
}

type Registry struct {
	endpoints map[string]EndpointSpec

	AllowOverride bool
}

func NewRegistry() *Registry {
	return &Registry{endpoints: map[string]EndpointSpec{}}
}

func (r *Registry) LoadFromFile(filePath, protoPackage string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	ctx := SourceInfo{
		Filename:     filePath,
		ProtoPackage: protoPackage,
	}

	switch filepath.Ext(filePath) {
	case ".json":
		return r.loadJSON(file, ctx)
	case ".yaml", ".yml":
		return r.loadYAML(file, ctx)
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

	config := &Config{}
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

	config := &Config{}
	if err := protojson.Unmarshal(content, config); err != nil {
		return fmt.Errorf("failed to unmarshal json file: %s", err)
	}

	return r.processConfig(config, src)
}

func (r *Registry) processConfig(config *Config, src SourceInfo) error {
	if config.Gateway == nil {
		return nil
	}

	for _, endpoint := range config.Gateway.GetEndpoints() {
		if err := validateBinding(endpoint); err != nil {
			return err
		}

		if existingBinding, ok := r.endpoints[endpoint.Selector]; ok && !r.AllowOverride {
			return fmt.Errorf(
				"conflicting binding for %q: both %q and %q contain bindings for this selector",
				endpoint.Selector, src.Filename, existingBinding.SourceInfo.Filename)
		}

		r.endpoints[endpoint.Selector] = EndpointSpec{
			Binding:    endpoint,
			SourceInfo: src,
		}
	}

	return nil
}

func (r *Registry) LookupBinding(selector string) (EndpointSpec, bool) {
	result, ok := r.endpoints[selector]
	return result, ok
}

func validateBinding(endpoint *EndpointBinding) error {
	if !selectorPattern.MatchString(endpoint.Selector) {
		return fmt.Errorf("invalid selector: %q", endpoint.Selector)
	}

	if endpoint.Body != "" && endpoint.Body != "*" && !selectorPattern.MatchString(endpoint.Selector) {
		return fmt.Errorf("invalid body selector for %q: %s", endpoint.Selector, endpoint.Body)
	}

	for _, binding := range endpoint.AdditionalBindings {
		if binding.Body != "" && binding.Body != "*" && !selectorPattern.MatchString(endpoint.Selector) {
			return fmt.Errorf("invalid body selector %q: %s", endpoint.Selector, endpoint.Body)
		}
	}

	return nil
}
