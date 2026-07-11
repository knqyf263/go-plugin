package gen

import (
	"bytes"
	"google.golang.org/protobuf/compiler/protogen"
	vtgenerator "github.com/planetscale/vtprotobuf/generator" // Confirme o import/alias utilizado no seu projeto
)

func (gg *Generator) generateVTFile(f *fileInfo) {
	// 1. Cria um plugin "dummy" em memória usando o request original para não sujar o plugin real
	dummyPlugin, err := protogen.Options{}.New(gg.plugin.Request)
	if err != nil {
		gg.plugin.Error(err)
		return
	}

	// 2. Inicializa o gerador vtprotobuf apontando para o plugin em memória
	cfg := &vtgenerator.Config{}
	featureNames := []string{"marshal", "unmarshal", "size"}
	localVtGen, err := vtgenerator.NewGenerator(dummyPlugin, featureNames, cfg)
	if err != nil {
		gg.plugin.Error(err)
		return
	}

	// 3. Processa e gera os arquivos dentro do dummyPlugin
	localVtGen.Generate()

	// 4. Extrai da resposta em memória apenas o arquivo que corresponde a este ciclo
	targetFilename := f.GeneratedFilenamePrefix + "_vtproto.pb.go"
	var vtContent []byte
	for _, rf := range dummyPlugin.Response().File {
		if rf.GetName() == targetFilename {
			vtContent = []byte(rf.GetContent())
			break
		}
	}

	// Se não gerou nada para este arquivo, abortamos silenciosamente (comportamento original)
	if len(vtContent) == 0 {
		return
	}

	// 5. Pula o cabeçalho original gerado
	if idx := bytes.Index(vtContent, []byte("import ")); idx != -1 {
		vtContent = vtContent[idx:]
	}

	// 6. Escreve no plugin REAL o seu cabeçalho customizado e o conteúdo interceptado
	g := gg.plugin.NewGeneratedFile(targetFilename, f.GoImportPath)
	gg.generateHeader(g, f)
	g.Write(vtContent)
}
