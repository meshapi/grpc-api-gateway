package main

import (
	"flag"

	"github.com/meshapi/grpc-rest-gateway/internal/codegen/gengateway"
)

func prepareOptions() *gengateway.Options {
	generatorOptions := gengateway.DefaultOptions()

	flag.StringVar(
		&generatorOptions.RegisterFunctionSuffix, "register_func_suffix", generatorOptions.RegisterFunctionSuffix,
		"used to construct names of generated Register*<Suffix> methods.")

	flag.BoolVar(
		&generatorOptions.UseHTTPRequestContext, "request_context", generatorOptions.UseHTTPRequestContext,
		"determine whether to use http.Request's context or not.")

	flag.BoolVar(
		&generatorOptions.AllowDeleteBody, "allow_delete_body", generatorOptions.AllowDeleteBody,
		"unless set, HTTP DELETE methods may not have a body")

	flag.Var(
		&generatorOptions.RepeatedPathParameterSeparator, "repeated_path_param_separator",
		"configures how repeated fields should be split. Allowed values are `csv`, `pipes`, `ssv` and `tsv`.")

	flag.BoolVar(
		&generatorOptions.AllowPatchFeature, "allow_patch_feature", generatorOptions.AllowPatchFeature,
		"determines whether to use PATCH feature involving update masks (using google.protobuf.FieldMask).")

	flag.BoolVar(
		&generatorOptions.OmitPackageDoc, "omit_package_doc", generatorOptions.OmitPackageDoc,
		"if true, no package comment will be included in the generated code")

	flag.BoolVar(
		&generatorOptions.Standalone, "standalone", generatorOptions.Standalone,
		"generates a standalone gateway package, which imports the target service package")

	generatorOptions.DescriptorRegistryOptions.AddFlags(flag.CommandLine)

	flag.Parse()

	return &generatorOptions
}
