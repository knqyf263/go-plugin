package main

import (
	"flag"
	"log"
	"os"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/knqyf263/go-plugin/gen"
)

func main() {
	//b, err := io.ReadAll(os.Stdin)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//if err = os.WriteFile("stdin.txt", b, os.ModePerm); err != nil {
	//	log.Fatal(err)
	//}
	//return
	if os.Getenv("GO_PLUGIN_DEBUG") != "" {
		input, err := os.Open("stdin.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer input.Close()
		os.Stdin = input
	}

	var flags flag.FlagSet
	protogen.Options{ParamFunc: flags.Set}.Run(func(plugin *protogen.Plugin) error {
		g, err := gen.NewGenerator(plugin)
		if err != nil {
			return err
		}

		for _, f := range plugin.Files {
			if !f.Generate {
				continue
			}
			g.GenerateFiles(f)
		}

		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		return nil
	})
}
