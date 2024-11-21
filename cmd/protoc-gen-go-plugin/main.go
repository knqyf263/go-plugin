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
	protogen.Options{ParamFunc: flags.Set}.Run(func(plugin *protogen.Plugin) error {
		g, err := gen.NewGenerator(plugin)
		if err != nil {
			return err
		}

		for _, f := range plugin.Files {
			if !f.Generate {
				continue
			}
			g.GenerateFiles(f, gen.Options{DisablePBGen: *disablePbGen})
		}

		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		return nil
	})
}
