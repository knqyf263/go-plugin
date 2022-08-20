package gen

import (
	"bytes"
)

func (gg *Generator) generateVTFile(f *fileInfo) {
	// It is a bit tricky, but necessary for workaround.
	// Generates a vtprotobuf file as a dummy file, modifies and writes that to the real file.
	filename := f.GeneratedFilenamePrefix + "_vt_dummy.pb.go"
	g := gg.plugin.NewGeneratedFile(filename, f.GoImportPath)
	gg.vtgen.GenerateFile(g, f.File)

	b, err := g.Content()
	if err != nil {
		gg.plugin.Error(err)
		return
	}

	// Skip the generated header
	if idx := bytes.Index(b, []byte("import ")); idx != -1 {
		b = b[idx:]
	}

	// Do not generate the dummy file
	g.Skip()

	// Write the replaced content
	filename = f.GeneratedFilenamePrefix + "_vtproto.pb.go"
	g = gg.plugin.NewGeneratedFile(filename, f.GoImportPath)
	gg.generateHeader(g, f)
	g.Write(b)
}
