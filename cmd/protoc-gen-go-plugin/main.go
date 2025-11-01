package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/knqyf263/go-plugin/gen"
)

func main() {
	var flags flag.FlagSet
	disablePbGen := flags.Bool("disable_pb_gen", false, "disable .pb.go generation")
	wasmPackage := flags.String("wasm_package", "github.com/knqyf263/go-plugin/wasm", "override package that provide wasm memory management")
	useGoPluginKnownTypes := flags.Bool("goplugin_known_types", true, "use go-plugin known types")
	protogen.Options{ParamFunc: flags.Set}.Run(func(plugin *protogen.Plugin) error {
		opts := gen.Options{
			UseGoPluginKnownTypes: *useGoPluginKnownTypes,
			DisablePBGen:          *disablePbGen,
			WasmPackage:           *wasmPackage,
		}
		g, err := gen.NewGenerator(plugin, opts)
		if err != nil {
			return err
		}

		for _, f := range plugin.Files {
			if !f.Generate {
				continue
			}
			g.GenerateFiles(f, opts)
		}

		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		return nil
	})
}
