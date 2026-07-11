package gen

import (
	"bytes"
	"google.golang.org/protobuf/compiler/protogen"
	vtgenerator "github.com/planetscale/vtprotobuf/generator" // Confirme o import/alias utilizado no seu projeto
)

func (gg *Generator) generateVTFile(f *fileInfo) {
	// It is a bit tricky, but necessary for workaround.
	// We create a dummyPlugin in memory.
	// Create the VTfile with it and then update its header
	dummyPlugin, err := protogen.Options{}.New(gg.plugin.Request)
	if err != nil {
		gg.plugin.Error(err)
		return
	}

	cfg := &vtgenerator.Config{}
	featureNames := []string{"marshal", "unmarshal", "size"}
	localVtGen, err := vtgenerator.NewGenerator(dummyPlugin, featureNames, cfg)
	if err != nil {
		gg.plugin.Error(err)
		return
	}

	localVtGen.Generate()

	targetFilename := f.GeneratedFilenamePrefix + "_vtproto.pb.go"
	var vtContent []byte
	for _, rf := range dummyPlugin.Response().File {
		if rf.GetName() == targetFilename {
			vtContent = []byte(rf.GetContent())
			break
		}
	}

	if len(vtContent) == 0 {
		return
	}

	if idx := bytes.Index(vtContent, []byte("import ")); idx != -1 {
		vtContent = vtContent[idx:]
	}

	g := gg.plugin.NewGeneratedFile(targetFilename, f.GoImportPath)
	gg.generateHeader(g, f)
	g.Write(vtContent)
}
