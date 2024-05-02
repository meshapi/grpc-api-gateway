// Command protoc-gen-grpc-rest-gateway is a plugin for Google protocol buffer
// compiler to generate a reverse-proxy, which converts incoming RESTful
// HTTP/1 requests gRPC invocation.
// You rarely need to run this program directly. Instead, put this program
// into your $PATH with a name "protoc-gen-grpc-gateway" and run
//
//	protoc --grpc-rest-gateway_out=output_directory path/to/input.proto
//
// See README.md for more details.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/meshapi/grpc-rest-gateway/codegen/internal/descriptor"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/gengateway"
	"github.com/meshapi/grpc-rest-gateway/codegen/internal/genlog"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	showVersion := flag.Bool("version", false, "show version")
	logFile := flag.String("log_file", "", "path to the output log file")
	logLevel := genlog.LevelWarning
	flag.Var(&logLevel,
		"log_level", "sets the log level, levels: 'warning', 'info', 'trace' and 'silent' (default: warning).")

	generatorOptions := prepareOptions()
	registryOptions := descriptor.DefaultRegistryOptions()
	registryOptions.AddFlags(flag.CommandLine)
	flag.Parse()
	genlog.SetLevel(logLevel)

	if *showVersion {
		fmt.Printf("Version v0.1.0\n")
		os.Exit(0)
	}

	options := protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}

	options.Run(func(gen *protogen.Plugin) error {
		if *logFile != "" {
			writer, err := os.Create(*logFile)
			if err != nil {
				return fmt.Errorf("failed to create log file: %w", err)

			}
			defer writer.Close()

			genlog.Set(genlog.New(writer, logLevel))
		}

		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		descriptorRegistry := descriptor.NewRegistry(registryOptions)

		if err := descriptorRegistry.LoadFromPlugin(gen); err != nil {
			return fmt.Errorf("failed to prepare for generation: %w", err)
		}

		if unboundSpecs := descriptorRegistry.UnboundExternalHTTPSpecs(); len(unboundSpecs) > 0 {
			values := []string{}
			for _, spec := range unboundSpecs {
				values = append(values, fmt.Sprintf("%s (%s)", spec.Binding.Selector, spec.SourceInfo.Filename))
			}
			return fmt.Errorf("HTTP binding specifications without a matching selector: %s", strings.Join(values, ", "))
		}

		targets := make([]*descriptor.File, len(gen.Request.FileToGenerate))
		for index, file := range gen.Request.FileToGenerate {
			target, err := descriptorRegistry.LookupFile(file)
			if err != nil {
				return err
			}

			targets[index] = target
		}

		generator := gengateway.New(descriptorRegistry, *generatorOptions)
		responseFiles, err := generator.Generate(targets)
		for _, file := range responseFiles {
			generatedFile := gen.NewGeneratedFile(file.GetName(), protogen.GoImportPath(file.GoPkg.Path))
			if _, err := generatedFile.Write([]byte(file.GetContent())); err != nil {
				return fmt.Errorf("error writing generated file content: %w", err)
			}
		}

		if err != nil {
			return err
		}

		return nil
	})
}
