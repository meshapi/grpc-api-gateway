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

	"github.com/meshapi/grpc-rest-gateway/internal/codegen/gengateway"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	showVersion := flag.Bool("version", false, "show version")
	generatorOptions := prepareOptions()

	if *showVersion {
		fmt.Printf("Version v0.1.0\n")
		os.Exit(0)
	}

	options := protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}

	options.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		generator := gengateway.New(generatorOptions)
		if err := generator.LoadFromPlugin(gen); err != nil {
			grpclog.Fatalf("failed to prepare for generation: %s", err)
		}

		grpclog.Infof("received request for %s", gen.FilesByPath)
		return nil
	})
}